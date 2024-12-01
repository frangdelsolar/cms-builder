# Git checkout master and read version
# make sure the files update to the next version
import logging
import colorlog

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
    

if __name__ == "__main__":
    logger.info("Version Validator")