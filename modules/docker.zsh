# modules/docker.zsh
# Docker utilities for container management and cleanup

# Commands:
#   containers       List running containers (ID and image)
#   murderdocker     Aggressively prune all Docker resources

alias containers="docker ps --format \"table {{.ID}} \t{{.Image}}\""

# Docker is being annoying, let's just burn it to the ground.
murderdocker() {
  docker volume prune
  docker container prune --filter "until=24h"
  docker image prune -a --filter "until=24h"
  docker network prune
}
