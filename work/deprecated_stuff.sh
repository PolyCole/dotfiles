# ************************************
#          Protocol Buffers
# ************************************
alias pblint="pblint "
alias pbcheck="pbcheck "
alias pbbuild="pbbuild "

# Lints either all proto bufs, or a specified pb package.
function pblint() {
  if [ $# -eq 0 ]; then
		docker-compose run defs -- lint_protos;
	else
		docker-compose run defs -- lint_protos "$@";
	fi;
}

# Checks either all proto bufs for breaking changes, or a specified pb package.
function pbcheck() {
  if [ $# -eq 0 ]; then
		saml2aws exec --exec-profile monolith -- docker-compose run defs -- break_check_protos;
	else
		saml2aws exec --exec-profile monolith -- docker-compose run defs -- break_check_protos "$@";
	fi;
}

# Builds either all proto bufs, or a specified pb package.
function pbbuild() {
  if [ $# -eq 0 ]; then
		docker-compose run defs -- build_protos;
	else
		docker-compose run defs -- build_protos "$@";
	fi;
}

# ************************************
#       Kubernetes Configuration
# ************************************

# DEPRECATED I THINK
# Builds and deploys a local branch to staging.
argo-deploy-stage-local() {
  kubeconfig-stage

  currentRev=$(git describe --always)
  repo=$(basename `git rev-parse --show-toplevel`)
  commitMessage=$(git log -1 --pretty=%B)

  imageTag="${1}"

  if [ -z "$imageTag" ]; then
	   imageTag="git-$(git describe --always)" \
		   || error "Could not get git revision"
  fi

  printf "****************************************************\n"
  printf "~~Deploying revision:~~\n"
  printf "$currentRev\n"
  printf "$commitMessage\n"
  printf "\n~~Deploying to service:~~\n"
  printf "$repo\n"
  printf "\n~~Deploying ecr image:~~\n"
  printf "$imageTag\n"
  printf "****************************************************\n"

  gradle clean build
  saml2aws exec --exec-profile monolith -- ./scripts/ecr.sh $repo $currentRev

  ./scripts/deploy.sh staging $imageTag
}

# DEPRECATED I THINK
# Deploying to production via Argo.
argo-deploy-prod(){
  kubeconfig-prod
  ./scripts/deploy.sh production
}