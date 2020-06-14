import os
import subprocess

# Executes a command.
def execute(command, cwd = os.getcwd()):
    arguments = command.split()
    print("$ {}".format(command))

    process = subprocess.Popen(args=arguments, cwd=cwd)
    exit_code = process.wait()

    if exit_code != 0:
        exit(exit_code)