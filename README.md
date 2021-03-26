# Cole Polyak's Dotfiles :record_button:
This repo hosts the dotfiles I accrue over the years. There are a couple of purposes this repository serves:
1. Ensures that dotfiles and config files are backed up regularly in case I need to get up to speed config-wise on a new machine.
2. Ensures that helpful functions or aliases I come across are persisted across devices, and that I always have access to them moving forward.
3. To take the config backup process out of my hands completely, and place it into the far more reliable hands of the `crond` 😉

## Note: crontab can be kind of finicky.
Thanks to [Jesse Spevack](https://github.com/jesse-spevack/) for figuring out how to get cached GitHub credentials working inside of crontab.

Step 1: define the `PATH` for crontab, which is doable with this [script](https://github.com/ssstonebraker/braker-scripts/blob/master/working-scripts/add_current_shell_and_path_to_crontab.sh) found via [stack overflow](https://stackoverflow.com/questions/2388087/how-to-get-cron-to-call-in-the-correct-paths).

Step 2: Ensure access to the Github repo is done using SSH, which is succinctly summarized by [this gist](https://gist.github.com/developius/c81f021eb5c5916013dc)

Step 3: Profit :)
