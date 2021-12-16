from pathlib import Path

filename = "count-log-events/instance-events.txt"

my_file = Path(filename)
if my_file.is_file():
    # file exists

    line_number = 0
    error_found = 0

    with open(filename) as file:
        for line in file:
            line_number += 1
            if "error" in line:
                if "dp-download-service" in line:
                    # As at 16th December 2021 the is bot fully used in the test stack, so not all of its dependencies
                    # have been set up and to avoid false positive error reporting we ignore any of its errors.
                    continue
                if error_found == 0:
                    error_found = 1
                    print("\nFound error(s) in: ", filename, "\n")
                print("line: ", line_number, "\n  ", line.lstrip())
