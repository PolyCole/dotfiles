# machines/personal/modules/woog.zsh — Wooglin API & bot project aliases

# Commands:
#   runserver     - start the app (python manage.py runserver)
#   make          - make the migrations (python manage.py makemigrations)
#   migrate       - apply migrations (python manage.py migrate)
#   test          - run tests (python manage.py test)
#   shell         - open app shell (python manage.py shell)
#   refresh-reqs  - reload requirements.txt
#   woogapi       - cd to wooglin-api and activate venv
#   woogbot       - cd to wooglin-bot and activate venv

# Django Stuff
alias runserver="python manage.py runserver"
alias migrate="python manage.py migrate"
alias test="python manage.py test"
alias shell="python manage.py shell"
alias refresh-reqs="rm requirements.txt && pip freeze >> requirements.txt"

woog-commands() {
  print "\n*******************************************"
  print "runserver -- starts the app"
  print "make -- makes the migrations"
  print "migrate -- applies migrations"
  print "test -- runs test"
  print "shell -- opens app shell"
  print "refresh-reqs -- re-loads requirements.txt"
  print "******************************************\n"

}

woogapi() {
  cd ~/repos/wooglin-api
  source venv/bin/activate
}

woogbot() {
  cd ~/repos/wooglin-bot
  source venv/bin/activate
}
