# machines/ibotta/modules/deploy.zsh — Deployment workflow functions

# Commands:
#   build_and_promote <service> <git-hash>   - Build and promote a microservice
#   promote <service> <git-hash> <env>       - Promote a microservice to an environment
#   deploy_production <git-hash> <service>   - Deploy a microservice to production
#   mono-deploy-stage <branch>               - Deploy branch to monolith staging (receipts)
#   deploy_mono_stage <instance> <branch>    - Deploy branch to a specific monolith staging instance
#   reset_mono_stage <instance>              - Tear down and recreate a monolith staging instance
#   thanks_for_your_contribution_detekt      - Run detekt, commit fixes, and push

# ---------------------------------------------------------------------------
# Microservice Deployments
# ---------------------------------------------------------------------------
function build_and_promote() {
  if [ $# -ne 2 ]; then
    print "Please specify the service and commit hash using: \n"
    print "build_and_promote [service] git-[commit hash]"
  else
    echo "************** Building **************"
    gradle clean build

    echo "************** Deploying **************"
    echo running saml2aws exec --exec-profile monolith -- $REPO_BASE_PATH/travis-shared-config/deployment/publish_and_promote.sh $1 $2
    saml2aws exec --exec-profile monolith -- $REPO_BASE_PATH/travis-shared-config/deployment/publish_and_promote.sh $1 $2
  fi
}

function promote() {
  if [ $# -ne 3 ]; then
    print "Please specify the service and commit hash using: \n"
    print "promote [service] git-[commit hash] [environment]"
  else
    echo saml2aws exec --exec-profile monolith -- $REPO_BASE_PATH/travis-shared-config/deployment/promote.sh $2 $3 $1
    saml2aws exec --exec-profile monolith -- $REPO_BASE_PATH/travis-shared-config/deployment/promote.sh $2 $1 $3
  fi
}

alias deploy_production="deploy_production "

function deploy_production() {
  if [ $# -lt 2 ]; then
    echo "Improper command format. Please use:"
    echo "\tdeploy_production git-<COMMIT_HASH> <SERVICE_NAME>"
  else
    saml2aws exec --exec-profile monolith -- $REPO_BASE_PATH/travis-shared-config/deployment/promote.sh $1 production $2
  fi;
}

# ---------------------------------------------------------------------------
# Monolith Deployment
# ---------------------------------------------------------------------------

# Deploying published branch to monolith staging
mono-deploy-stage() {
  branchName="${1}"

  if [ -z "$branchName" ]; then
	   error "Must specify a branch name to deploy: mono-deploy-stage <branchName>"
  fi
  saml2aws exec --exec-profile staging -- ibotta_cli application deploy api-server name:receipts -r $branchName
}

# Deploys the specified branch to the specified monolith instance.
function deploy_mono_stage() {
  if [ $# -ne 2 ]; then
    print "Please specify the instance and branchname using: \n"
    print "deploy_mono_stage [instance] [branch]\n"
  else
    echo running: saml2aws exec --exec-profile staging -- ibotta_cli application deploy api-server name:$1 -r $2
    saml2aws exec --exec-profile staging -- ibotta_cli application deploy api-server name:$1 -r $2
  fi
}

# Pulls down and boots up the specified monolith staging instance.
function reset_mono_stage() {
  if [ $# -ne 1 ]; then
    print "Please specify the instance name using: \n"
    print "reset_mono_stage [instance]\n"
  else
    cd $REPO_BASE_PATH/infrastructure
    print "\n\nTearing and Rolling instance: $1\n\n"
    saml2aws exec --exec-profile staging -- ibotta_cli sfn destroy $1
    saml2aws exec --exec-profile staging -- ibotta_cli sfn create $1 -f sparkleformation/development.rb
  fi
}

# ---------------------------------------------------------------------------
# Detekt Workflow
# ---------------------------------------------------------------------------
function thanks_for_your_contribution_detekt() {
  print "\n\n**Detekt is an A1 linter**\n\n"
  ./gradlew detekt
  print "\n\n**I love every part of it**\n\n"
  git add -A
  print "\n\n**Especially that it flunks builds for arbitrary things**\n\n"
  git commit -m "Submitting to the indomitable will of Detekt."
  print "\n\n**That it then fixes on its own!**\n\n"
  git push
}
