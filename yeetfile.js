const image = "ghcr.io/sapslaj/valkey-leader";

const dockerPlatforms = [
  "linux/amd64",
  "linux/arm64",
];

$`docker buildx build --platform=${dockerPlatforms.join(",")} -t ${image}:latest .`;

const dockerPush = yeet.getenv("DOCKER_PUSH");
const dockerTags = yeet.getenv("DOCKER_TAGS").split(",");

for (const dockerTag of dockerTags) {
  $`docker tag ${image}:latest ${image}:${dockerTag}`;
}

if (dockerPush === "true" || dockerPush === "latest") {
  $`docker push ${image}:latest`;
}
if (dockerPush === "true") {
  for (const dockerTag of dockerTags) {
    $`docker push ${image}:${dockerTag}`;
  }
}
