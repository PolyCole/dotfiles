# machines/ibotta/modules/oncall.zsh — On-call runbook commands

# Commands:
#   oncall-check    - Check production API server revisions
#   oncall-deploy   - Deploy API server to production
#   oncall-verify   - Verify production API server revisions
#   oncall-fix      - Deploy API server to production with --fix flag
#   oncall-migrate  - Run production database migrations
#   oncall-rollback - Rollback API server in production

# ---------------------------------------------------------------------------
# Monolith On-Call Runbook
# ---------------------------------------------------------------------------

alias oncall-check="saml-monolith ibotta_cli application revisions api-server 'environment:production'"
alias oncall-deploy="saml-monolith ibotta_cli application deploy_api_server_production"
alias oncall-verify="saml-monolith ibotta_cli application verify_revisions api-server 'role:api_server*'"
alias oncall-fix="saml-monolith ibotta_cli application deploy_api_server_production --fix"
alias oncall-migrate="saml-monolith ./bin/drake db:migration:prod"
alias oncall-rollback="saml-monolith ibotta_cli application rollback_api_server_production"
