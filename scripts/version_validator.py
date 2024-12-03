# Git checkout master and read version
# make sure the files update to the next version
import logging
import colorlog
from pathlib import Path

# Create a custom formatter with color and file/line information
formatter = colorlog.ColoredFormatter(
    '%(log_color)s%(levelname)-8s%(reset)s %(filename)s:%(lineno)d %(message)s'
)

# Create a logger instance
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

# Add a stream handler with the custom formatter
handler = logging.StreamHandler()
handler.setFormatter(formatter)
logger.addHandler(handler)


def check_current_version():
    "git checkout master"

def read_version_from_readme():
    readme_path = Path(__file__).parent.parent / "builder" / "README.md"

    logger.info("Read version from %s", readme_path)
    version = ""
    with readme_path.open() as f:
        for line in f:
            if line.startswith("#"):
                version = line.split("v")[1]
                break

    logger.info("Version: %s", version)
    return version

def update_version_in_readme(version):
    readme_path = Path(__file__).parent.parent / "builder" / "README.md"

    logger.info("Update version in %s", readme_path)
    with readme_path.open() as f:
        lines = f.readlines()

    with readme_path.open("w") as f:
        for ix, line in enumerate(lines):
            if ix == 0:
                appName, old_version = line.split("v")
                new_text = f"{appName}v{version}\n"
                f.write(new_text)
            else:
                f.write(line)

    logger.info("Updated version in %s", readme_path)

def read_version_from_builder_go():
    readme_path = Path(__file__).parent.parent / "builder" / "builder.go"

    logger.info("Read version from %s", readme_path)
    version = ""
    with readme_path.open() as f:
        for line in f:
            if line.startswith("const builderVersion"):
                version = line.split("=")[1].strip().replace("\"", "")
                break

    logger.info("Version: %s", version)
    return version    

def update_version_in_builder_go(version):
    readme_path = Path(__file__).parent.parent / "builder" / "builder.go"

    logger.info("Update version in %s", readme_path)
    with readme_path.open() as f:
        lines = f.readlines()

    with readme_path.open("w") as f:
        for line in lines:
            if line.startswith("const builderVersion"):
                f.write(f"const builderVersion = \"{version}\"\n")
            else:
                f.write(line)

    logger.info("Updated version in %s", readme_path)


if __name__ == "__main__":
    logger.info("Version Validator")

    read_version_from_readme()
    read_version_from_builder_go()
    new_version = "3.9.9"

    update_version_in_readme(new_version)
    update_version_in_builder_go(new_version)