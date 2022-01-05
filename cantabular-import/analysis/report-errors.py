from pathlib import Path

filename = "count-log-events/instance-events.txt"

line_number = 0
error_found = 0
download_error_count = 0

my_file = Path(filename)
if my_file.is_file():
    # file exists

    with open(filename) as file:
        for line in file:
            line_number += 1
            if "error" in line:
                if "dp-download-service" in line:
                    # As at 16th December 2021 this is not fully used in the test stack, so not all of its dependencies
                    # have been set up and to avoid false positive error reporting we ignore any of its errors.
                    download_error_count += 1
                    continue
                if error_found == 0:
                    error_found = 1
                    print("\nFound error(s) in: ", filename, "\n")
                print("line: ", line_number, "\n  ", line.lstrip())

if download_error_count > 0:
    print("    Had a download-service error count of: ", download_error_count, " -> ignore these !")

if error_found == 0:
    print("    No unexpected error(s) found\n")

# now look for `DATA RACE` in 'all-container-logs.txt'

filename = "tmp/all-container-logs.txt"

line_number = 0
error_found = 0

my_file = Path(filename)
if my_file.is_file():
    # file exists

    with open(filename) as file:
        for line in file:
            line_number += 1
            if "DATA RACE" in line:
                if error_found == 0:
                    error_found = 1
                    print("\nFound DATA RACE in: ", filename, "\n")
                print("line: ", line_number, "\n  ", line.lstrip())

if error_found == 0:
    print("    No DATA RACE's found\n")