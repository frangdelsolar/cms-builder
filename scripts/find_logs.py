import os
import re

class Colors:
    GREEN = "\033[0;32m"
    RED = "\033[0;31m"
    BLUE = "\033[0;34m"
    CYAN = "\033[0;36m"
    YELLOW = "\033[0;33m"
    WHITE = "\033[0m"
    BLACK = "\033[0;30m"


folder_path = "../builder"


def read_go_files(folder_path: str) -> list[str]:
    """
    Reads all .go files in the given folder and its subfolders and returns a list of the paths to the files.
    """
    files = []
    for root, dirs, filenames in os.walk(folder_path):
        for filename in filenames:
            if filename.endswith(".go"):
                files.append(os.path.join(root, filename))
    return files

def read_file_lines(file_path: str) -> None:
    """
    Reads all lines from the given file and prints them to the console.
    
    :param file_path: The path to the file to read from.
    """

    print(f"{Colors.BLACK}Checking: {file_path}{Colors.WHITE}")
    
    line_number = 0
    pattern = r"log\.(Debug|Error|Trace|Info|Warn)\(\)"

    with open(file_path, 'r') as file:
        lines = file.readlines()


    for line in lines:
        line_number += 1
        match = re.search(pattern, line)
        if match:
            log_level = match.group(1)

            print_log(file_path, line_number, line, log_level)


def print_log(file, line_number, line, log_level):

    color = {
        "Debug": Colors.GREEN,
        "Error": Colors.RED,
        "Trace": Colors.BLUE,
        "Info": Colors.CYAN,
        "Warn": Colors.YELLOW
    }.get(log_level, Colors.WHITE)  # Default

    print(f"{file.split('/')[-1]}:{line_number} -> {color}{line.strip()}{Colors.WHITE}")


def main():

    files = read_go_files(folder_path)

    for file in files:
        read_file_lines(file)

if __name__ == "__main__":
    main()