const image = "ghcr.io/sapslaj/valkey-leader";

const dockerPlatforms = [
  "linux/amd64",
  "linux/arm64",
];

const dockerTags = yeet.getenv("DOCKER_TAGS").split(",");

const dockerBuildTags = [
  `${image}:latest`,
];
for (const dockerTag of dockerTags) {
  if (dockerTag === "" || dockerTag === "latest") {
    continue;
  }
  dockerBuildTags.push(`${image}:${dockerTag}`);
}

const dockerPush = yeet.getenv("DOCKER_PUSH");

yeet.run(
  "docker",
  "buildx",
  "build",
  "--platform",
  dockerPlatforms.join(","),
  "-t",
  dockerBuildTags.join(","),
  ...(dockerPush === "true" ? ["--push"] : []),
  ".",
);
