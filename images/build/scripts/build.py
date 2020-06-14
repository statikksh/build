from os import environ as env, getcwd
from pathlib import Path

import json

import utils

# Constants and environment variables
REPOSITORY_URL = env.get('REPOSITORY')
SOURCES_DIRECTORY = Path(getcwd() + "/sources")

if __name__ == "__main__":
    print("Build started for {}.".format(REPOSITORY_URL))

    # Create a directory for the repository
    SOURCES_DIRECTORY.mkdir(parents = True, exist_ok = True)

    # Download the repository into the previous directory
    utils.execute("git clone {} {}".format(REPOSITORY_URL, str(SOURCES_DIRECTORY)))

    # Install repository source dependencies
    utils.execute("yarn", cwd=SOURCES_DIRECTORY)

    # Run the build script
    utils.execute("yarn run build", cwd=SOURCES_DIRECTORY)

    # TODO Send `public` directory contents to the static server.