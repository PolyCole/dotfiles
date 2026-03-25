# machines/ibotta/modules/aws.zsh — AWS and SAML2 authentication utilities

# Commands:
#   awslog    - Login to AWS via saml2aws (12hr session)
#   ecrlogin  - Authenticate Docker with Ibotta's ECR registry
#   ecrchk    - Run ECR image check script via monolith profile
#   aws-som   - Run aws commands under staging-synchronous-matching-lambda profile

alias awslog="saml2aws login --skip-prompt --force --session-duration=43200"
alias ecrlogin="(saml2aws exec --exec-profile monolith -- aws ecr get-login-password --region us-east-1) | docker login --password-stdin --username AWS 264606497040.dkr.ecr.us-east-1.amazonaws.com"
alias ecrchk="saml2aws exec --exec-profile monolith -- ~/ecrimagecheck.sh"
alias aws-som="saml2aws exec --exec-profile staging-synchronous-matching-lambda -- "
