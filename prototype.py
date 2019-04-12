"""
fastqube prototype
~~~~~~~~~~~~~~~~~~

:copyright: (c) 2019 Sander Bollen
:license: BSD-3-clause
"""
import argparse
import pathlib
import bitstruct

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("input", type=pathlib.Path)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("-c", "--compress", action="store_true")
    group.add_argument("-d", "--decompress", action="store_true")

    args = parser.parse_args()

    if args.compress:
        raise NotImplementedError
    elif args.decompress:
        raise NotImplementedError
    else:
        raise NotImplementedError
