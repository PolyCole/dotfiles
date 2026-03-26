# modules/advent.zsh
# Advent of Code template generation helpers

# Commands:
#   aoc [day]            - Generate Advent of Code template for given day
#   advent-of-code [day] - Same as aoc

alias aoc="advent-of-code "

function advent-of-code() {
  python $HOME/repos/AdventOfCode/template_aoc.py --day $1
}
