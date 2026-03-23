# Commands:
#   weather [location]   Fetch weather from wttr.in (defaults to Denver)
#   status-dog <code>    Open HTTP status dog for given status code

# Retrieves weather for specified location.
function weather() {
  if [ $# -ne 1 ]; then
    curl "https://wttr.in/Denver\?TuF0"
  else
    curl "https://wttr.in/$1\?TuF0"
  fi;
}

# Opens up to a specified HTTP Status Dog :)
function status-dog() {
  if [ $# -ne 1 ]; then
    print "Please specify the error code you're looking up using:\n"
    print "status-dog [errorcode]"
    exit 0
  fi;

	open "https://httpstatusdogs.com/$1"
}
