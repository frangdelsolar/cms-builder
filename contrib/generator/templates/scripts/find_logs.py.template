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
    BOLD = "\033[1m"
    ITALIC = "\033[3m"

log_colors = {
    "Debug": Colors.GREEN,
    "Error": Colors.RED,
    "Trace": Colors.BLUE,
    "Info": Colors.CYAN,
    "Warn": Colors.YELLOW
}

results_count = {
    "Debug": 0,
    "Error": 0,
    "Trace": 0,
    "Info": 0,
    "Warn": 0
}


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

def find_logs(file_path: str) -> None:
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
            results_count[log_level] += 1


def print_log(file, line_number, line, log_level):
    """
    Prints a log message with the given log level to the console.

    :param file: The file where the log message was found.
    :param line_number: The line number of the log message in the file.
    :param line: The line containing the log message.
    :param log_level: The log level of the log message.
    """
    color = log_colors.get(log_level, Colors.WHITE)  # Default

    print(f"{file.split('/')[-1]}:{line_number} -> {color}{line.strip()}{Colors.WHITE}")


def present_results():
    """
    Prints the results of the log search to the console.
    """
    
    print("\n\nResults:")
    for level, count in results_count.items():
        color = log_colors.get(level, Colors.WHITE)  # Default
        italic = Colors.ITALIC if count > 0 else ""
        print(f"{color}{Colors.BOLD}{italic}{level}: {count}")
    
    # Have a nice banner saying please review the logs and keep them to a minimum
    print(f"{Colors.WHITE}\n**************************************************")
    print(f"{Colors.CYAN}{Colors.BOLD}Please review the logs and keep them to a minimum.")
    print(f"{Colors.WHITE}**************************************************")

def get_files_from_env():

    """
    Tries to get the list of changed files from the environment variable CHANGED_FILES
    and returns it as a list of file paths relative to the current working directory.

    If the variable is not set, the method returns an empty list.
    """
    files = os.getenv("CHANGED_FILES")
    output = []
    if files is not None:
        files = files.split(" ")

        for file in files:
            output.append("../"+file)

    return output

def main():

    """
    Main entry point of the script.

    This method first tries to get changed files from the environment variable
    CHANGED_FILES. If it's not set, it reads all .go files from the ../builder
    folder and uses those for the log search.

    It then searches for log statements in the given files and prints the
    results to the console.
    """
    files = []

    try:
        files = get_files_from_env()
    except Exception as e:
        print(f"Error: {e}")

    if len(files) == 0 or files is None:
        folder_path = "../builder"
        files = read_go_files(folder_path)

    print(files)

    for file in files:
        find_logs(file)

    present_results()

if __name__ == "__main__":
    main()