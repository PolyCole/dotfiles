# Cole Polyak's Dotfiles :record_button:
The repo hosting my dotfiles, in case I quickly need to get back up to speed on a new machine.

As an added bonus, calling the backup_dotfiles script can be handled by cron to ensure all changes are periodically stashed.

## Note: crontab can be kind of finicky.
Thanks to [Jesse Spevack](https://github.com/jesse-spevack/) for figuring out how to get cached GitHub credentials working inside of crontab.

Step 1: define the `PATH` for crontab, which is doable with this [script](https://github.com/ssstonebraker/braker-scripts/blob/master/working-scripts/add_current_shell_and_path_to_crontab.sh) found via [stack overflow](https://stackoverflow.com/questions/2388087/how-to-get-cron-to-call-in-the-correct-paths).

Step 2: Ensure access to the Github repo is using SSH, which is succinctly summarized by [this gist](https://gist.github.com/developius/c81f021eb5c5916013dc)

Step 3: Profit :) 
