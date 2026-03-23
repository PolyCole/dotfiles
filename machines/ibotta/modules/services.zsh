# machines/ibotta/modules/services.zsh — Microservice & repo navigation aliases

# Commands:
#   bvs       - cd to barcode-verification-service
#   wtc       - cd to walmart-tc-service
#   ros       - cd to receipt-ocr-service
#   som       - cd to synchronous-matching-lambda
#   trs       - cd to transaction-return-service
#   rdc       - cd to receipt-data-coordinator
#   prs       - cd to purchase-rating-service
#   acs       - cd to auto-credit-service
#   earnings  - cd to earnings-service
#   drake     - run drake via saml2aws monolith profile
#   monolith  - cd to Ibotta monolith repo
#   rspec     - run rspec via drake/saml2aws
#   catalog   - cd to ibotta-catalog
#   tk        - cd to tech-knowledge
#   hub       - cd to enablement-data-hub
#   portal    - cd to enablement-data-portal
#   cha       - cd to consumer-hub-adapter
#   scripts   - show package.json scripts via jq
#   nx        - shortcut for npx nx

# ---------------------------------------------------------------------------
# Monolith
# ---------------------------------------------------------------------------
alias drake="saml2aws exec --exec-profile monolith -- bin/drake"
alias monolith="$REPO_BASE_PATH/Ibotta"
alias rspec="saml2aws exec --exec-profile monolith -- bin/drake spring rspec"

# ---------------------------------------------------------------------------
# DevEx repos
# ---------------------------------------------------------------------------
alias nx="npx nx"
alias catalog="$REPO_BASE_PATH/ibotta-catalog"
alias tk="$REPO_BASE_PATH/tech-knowledge"
alias hub="$REPO_BASE_PATH/enablement-data-hub"
alias portal="$REPO_BASE_PATH/enablement-data-portal"
alias cha="$REPO_BASE_PATH/consumer-hub-adapter"
alias scripts="cat package.json| jq .scripts"

# ---------------------------------------------------------------------------
# Microservices
# ---------------------------------------------------------------------------
alias bvs="$REPO_BASE_PATH/barcode-verification-service"
alias wtc="$REPO_BASE_PATH/walmart-tc-service"
alias ros="$REPO_BASE_PATH/receipt-ocr-service"
alias som="$REPO_BASE_PATH/synchronous-matching-lambda"
alias trs="$REPO_BASE_PATH/transaction-return-service"
alias rdc="$REPO_BASE_PATH/receipt-data-coordinator"
alias prs="$REPO_BASE_PATH/purchase-rating-service"
alias acs="$REPO_BASE_PATH/auto-credit-service"
alias earnings="$REPO_BASE_PATH/earnings-service"
