package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/valkey-io/valkey-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/sapslaj/valkey-leader/pkg/env"
)

func main() {
	var leading atomic.Bool
	leading.Store(false)

	mainLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("error building Kubernetes config", slog.Any("error", err))
		os.Exit(1)
	}
	client := clientset.NewForConfigOrDie(config)

	clusterName := env.MustGet[string]("CLUSTER_NAME")
	namespace := env.MustGet[string]("NAMESPACE")
	podIP := env.MustGet[string]("POD_IP")
	podName := env.MustGet[string]("POD_NAME")
	serviceName := env.MustGet[string]("SERVICE_NAME")
	leaderLeaseName := env.MustGetDefault("LEADER_LEASE_NAME", clusterName)

	mainLogger = mainLogger.With(
		slog.String("cluster_name", clusterName),
		slog.String("namespace", namespace),
		slog.String("pod_ip", podIP),
		slog.String("pod_name", podName),
		slog.String("service_name", serviceName),
		slog.String("leader_lease_name", leaderLeaseName),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add cluster label to current pod
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		mainLogger.Error("failed to get current pod", slog.Any("error", err))
		os.Exit(1)
	}

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels["valkey.sapslaj.cloud/cluster"] = clusterName

	_, err = client.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	if err != nil {
		mainLogger.Error("failed to update pod labels", slog.Any("error", err))
		os.Exit(1)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		mainLogger.Info("received termination, signaling shutdown")
		leading.Store(false)
		cancel()
	}()

	go func() {
		for {
			logger := mainLogger.With()
			select {
			case <-time.After(15 * time.Second):
				if leading.Load() {
					continue
				}

				// Find the primary pod by looking for instance-role=primary label
				pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: "valkey.sapslaj.cloud/instance-role=primary",
				})
				if err != nil {
					logger.Error("failed to list pods", slog.Any("error", err))
					continue
				}

				if len(pods.Items) == 0 {
					logger.Warn("no primary pod found, retrying in 15 seconds")
					continue
				}

				primaryPod := pods.Items[0]
				primaryIP := primaryPod.Status.PodIP
				if primaryIP == "" {
					logger.Warn("primary pod has no IP address, retrying in 15 seconds")
					continue
				}

				logger.Info("found primary pod", slog.String("primary_pod", primaryPod.Name), slog.String("primary_ip", primaryIP))

				// Connect to local Valkey and configure replication
				valkeyClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"localhost:6379"}})
				if err != nil {
					logger.Error("failed to create Valkey client", slog.Any("error", err))
					continue
				}

				err = valkeyClient.Do(ctx, valkeyClient.B().Replicaof().Host(primaryIP).Port(6379).Build()).Error()
				valkeyClient.Close()
				if err != nil {
					logger.Error("failed to configure replication", slog.Any("error", err))
					continue
				}

				logger.Info("configured replication", slog.String("primary_ip", primaryIP))

				// Add replica label to current pod
				currentPod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
				if err != nil {
					logger.Error("failed to get current pod", slog.Any("error", err))
					continue
				}

				if currentPod.Labels == nil {
					currentPod.Labels = make(map[string]string)
				}
				currentPod.Labels["valkey.sapslaj.cloud/instance-role"] = "replica"

				_, err = client.CoreV1().Pods(namespace).Update(ctx, currentPod, metav1.UpdateOptions{})
				if err != nil {
					logger.Error("failed to update pod labels", slog.Any("error", err))
					continue
				}

				logger.Info("updated pod with replica label")

			case <-ctx.Done():
				logger.InfoContext(ctx, "context canceled")
				return
			}
		}
	}()

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaderLeaseName,
			Namespace: namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podIP,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				leading.Store(true)
				for leading.Load() {
					logger := mainLogger.With()
					select {
					case <-time.After(15 * time.Second):

					// Connect to local Valkey and promote to primary
					valkeyClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"localhost:6379"}})
					if err != nil {
						logger.Error("failed to create Valkey client", slog.Any("error", err))
						continue
					}

					err = valkeyClient.Do(ctx, valkeyClient.B().Replicaof().No().One().Build()).Error()
					valkeyClient.Close()
					if err != nil {
						logger.Error("failed to promote to primary", slog.Any("error", err))
						continue
					}

					logger.Info("promoted to primary")

					// Add primary label to current pod
					currentPod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
					if err != nil {
						logger.Error("failed to get current pod", slog.Any("error", err))
						continue
					}

					if currentPod.Labels == nil {
						currentPod.Labels = make(map[string]string)
					}
					currentPod.Labels["valkey.sapslaj.cloud/instance-role"] = "primary"

					_, err = client.CoreV1().Pods(namespace).Update(ctx, currentPod, metav1.UpdateOptions{})
					if err != nil {
						logger.Error("failed to update pod labels", slog.Any("error", err))
						continue
					}

					logger.Info("updated pod with primary label")

					case <-ctx.Done():
						logger.InfoContext(ctx, "context canceled")
						return
					}
				}
			},
			OnStoppedLeading: func() {
				mainLogger.Info("leader lost")
				leading.Store(false)
			},
			OnNewLeader: func(identity string) {
				mainLogger.Info("new leader elected", slog.String("identity", identity), slog.Bool("self", identity == podIP))
			},
		},
	})

}
