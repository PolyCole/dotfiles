#!/bin/bash
# filepath: ibotta-prs-by-week.sh

# Colors for output formatting
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
BOLD='\033[1m'
RESET='\033[0m'

# Default parameters
AUTHOR=""
WEEKS=4
STATE="all"
ORG="Ibotta"
REPO=""

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --author=*) AUTHOR="${1#*=}"; shift ;;
    --weeks=*) WEEKS="${1#*=}"; shift ;;
    --state=*) STATE="${1#*=}"; shift ;; # Can be open, closed, merged, all
    --org=*) ORG="${1#*=}"; shift ;;
    --repo=*) REPO="${1#*=}"; shift ;;
    --help) 
      echo "Usage: ibotta-prs-by-week.sh [options]"
      echo "Options:"
      echo "  --author=USERNAME   Filter PRs by author (default: your username)"
      echo "  --weeks=N           Show PRs from the last N weeks (default: 4)"
      echo "  --state=STATE       Filter by state: open, closed, merged, all (default: all)"
      echo "  --org=ORG           Specify organization (default: Ibotta)"
      echo "  --repo=REPO         Specify repository name only (default: all repos)"
      echo "  --help              Show this help message"
      exit 0
      ;;
    *) echo "Unknown parameter: $1"; exit 1 ;;
  esac
done

# Check if gh is installed
if ! command -v gh &> /dev/null; then
  echo "GitHub CLI (gh) is not installed. Please install it first."
  echo "See: https://cli.github.com/manual/installation"
  exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
  echo "You are not authenticated with GitHub CLI. Please run 'gh auth login' first."
  exit 1
fi

# If author is not specified, use the currently logged-in user
if [[ -z "$AUTHOR" ]]; then
  AUTHOR=$(gh api user | jq -r '.login')
  echo -e "${GRAY}No author specified, using current user: ${AUTHOR}${RESET}"
fi

# Calculate date from N weeks ago (cross-platform)
START_DATE=$(date -v-${WEEKS}w +%Y-%m-%d 2>/dev/null || date --date="$WEEKS weeks ago" +%Y-%m-%d)

echo -e "${BLUE}Fetching PRs created after ${BOLD}${START_DATE}${RESET}${BLUE}...${RESET}"

# Create a temporary file for storing PR data
TMP_FILE=$(mktemp)

# Function to fetch PRs for a specific repo
fetch_repo_prs() {
  local repo="$1"
  echo -e "${GRAY}Fetching PRs for ${repo}...${RESET}"
  
  # Determine the PR state filter for gh cli
  local state_filter="all"
  if [[ "$STATE" == "open" || "$STATE" == "closed" || "$STATE" == "merged" ]]; then
    state_filter="$STATE"
  fi
  
  # Use gh pr list which is more reliable than the API
  gh pr list --repo "${ORG}/${repo}" \
             --author "$AUTHOR" \
             --state "$state_filter" \
             --json number,title,url,createdAt,state,mergedAt,closedAt,additions,deletions,changedFiles \
             --limit 100 | jq -c '.[]' | while read -r pr; do
    
    PR_NUM=$(echo "$pr" | jq -r '.number')
    PR_TITLE=$(echo "$pr" | jq -r '.title')
    PR_URL=$(echo "$pr" | jq -r '.url')
    PR_CREATED=$(echo "$pr" | jq -r '.createdAt')
    PR_STATE=$(echo "$pr" | jq -r '.state')
    PR_MERGED=$(echo "$pr" | jq -r '.mergedAt')
    PR_ADDITIONS=$(echo "$pr" | jq -r '.additions')
    PR_DELETIONS=$(echo "$pr" | jq -r '.deletions')
    PR_CHANGED_FILES=$(echo "$pr" | jq -r '.changedFiles')
    
    # Skip PRs created before our start date
    PR_CREATED_TS=$(date -d "$PR_CREATED" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$PR_CREATED" +%s 2>/dev/null)
    START_DATE_TS=$(date -d "$START_DATE" +%s 2>/dev/null || date -j -f "%Y-%m-%d" "$START_DATE" +%s 2>/dev/null)
    
    if [[ $PR_CREATED_TS -lt $START_DATE_TS ]]; then
      continue
    fi
    
    # Calculate week start (Sunday) - cross-platform
    PR_DATE=$(date -d "$PR_CREATED" +%Y-%m-%d 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$PR_CREATED" +%Y-%m-%d 2>/dev/null)
    DOW=$(date -d "$PR_DATE" +%w 2>/dev/null || date -j -f "%Y-%m-%d" "$PR_DATE" +%w 2>/dev/null)
    WEEK_START=$(date -d "$PR_DATE -$DOW days" +%Y-%m-%d 2>/dev/null || date -j -v-${DOW}d -f "%Y-%m-%d" "$PR_DATE" +%Y-%m-%d 2>/dev/null)
    
    # Determine status
    if [[ -n "$PR_MERGED" && "$PR_MERGED" != "null" ]]; then
      STATUS="merged"
    elif [[ "$PR_STATE" == "CLOSED" || "$PR_STATE" == "closed" ]]; then
      STATUS="closed"
    else
      STATUS="open"
    fi
    
    # Skip PRs that don't match state filter
    if [[ "$STATE" != "all" ]]; then
      if [[ "$STATE" == "merged" && "$STATUS" != "merged" ]]; then
        continue
      elif [[ "$STATE" == "closed" && "$STATUS" != "closed" ]]; then
        continue
      elif [[ "$STATE" == "open" && "$STATUS" != "open" ]]; then
        continue
      fi
    fi
    
    # Write to temp file: WeekStart|RepoName|PRNum|Status|Title|URL|Additions|Deletions|Files
    echo "$WEEK_START|$repo|$PR_NUM|$STATUS|$PR_TITLE|$PR_URL|$PR_ADDITIONS|$PR_DELETIONS|$PR_CHANGED_FILES" >> "$TMP_FILE"
  done
}

# If specific repo provided, just query that one
if [[ -n "$REPO" ]]; then
  fetch_repo_prs "$REPO"
else
  # Otherwise, get list of repos in the org the user has access to
  echo -e "${GRAY}Fetching repositories in ${ORG}...${RESET}"
  REPOS=$(gh repo list "$ORG" --limit 100 --json name -q '.[].name')
  
  if [[ -z "$REPOS" ]]; then
    echo -e "${RED}No repositories found in ${ORG} that you have access to.${RESET}"
    exit 1
  fi
  
  # Process each repo
  for repo in $REPOS; do
    fetch_repo_prs "$repo"
  done
fi

# Check if we found any PRs
if [[ ! -s "$TMP_FILE" ]]; then
  echo -e "\n${RED}No PRs found matching your criteria.${RESET}"
  echo -e "${GRAY}Try adjusting the search parameters or check if you have access to the repositories.${RESET}"
  rm "$TMP_FILE"
  exit 0
fi

# Count total PRs
TOTAL=$(wc -l < "$TMP_FILE" | tr -d ' ')
echo -e "\n${GREEN}Found ${BOLD}$TOTAL${RESET}${GREEN} PRs${RESET}\n"

# Sort by week (newest first) and output with formatting
sort -r "$TMP_FILE" | awk -F'|' '
BEGIN {
  BLUE = "\033[0;34m";
  GREEN = "\033[0;32m";
  YELLOW = "\033[0;33m";
  RED = "\033[0;31m";
  GRAY = "\033[0;90m";
  CYAN = "\033[1;36m";
  BOLD = "\033[1m";
  RESET = "\033[0m";
}
{
  if (lastWeek != $1) {
    lastWeek = $1;
    printf "\n%sWeek of %s:%s\n", CYAN, $1, RESET;
  }
  
  # Format status with color
  if ($4 == "merged") {
    status = GREEN "merged" RESET;
  } else if ($4 == "closed") {
    status = RED "closed" RESET;
  } else {
    status = YELLOW "open" RESET;
  }
  
  printf "  %s%s%s #%s: %s%s%s (%s) +%s/-%s files:%s\n    %s\n", 
    GRAY, $2, RESET, $3, BOLD, $5, RESET, status, $7, $8, $9, $6;
}'

# Clean up
rm "$TMP_FILE"
echo -e "\n${BLUE}Done!${RESET}"