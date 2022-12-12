import os 
import random

functional_width = 40

def main():
    messages = get_messages()
    selected_message = messages[random.randint(0, len(messages) - 1)]
    print_message(get_lines(selected_message))

def get_messages():
    file = open(f"{os.environ['HOME']}/repos/dotfiles/personal/messages.txt", "r")

    arr = []
    for line in file:
        if line != '':
            arr.append(line.strip())
    
    return arr

def get_lines(message):
    words = message.split(" ")
    new_message = []
    cur_line = ""

    for word in words:
        if len(cur_line) + len(word) < 34:
            cur_line += f"{word} "
        else:
            new_message.append(cur_line)
            cur_line = f"{word} "

    new_message.append(cur_line)

    return new_message

def print_message(lines):
    for line in lines:
        buffer = ' '
        buffer *= (abs(len(line.strip()) - functional_width) // 2)
        print(f"{buffer}  {line.strip()}  {buffer}")

if __name__ == "__main__":
    main()

