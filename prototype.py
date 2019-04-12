"""
fastqube prototype
~~~~~~~~~~~~~~~~~~

:copyright: (c) 2019 Sander Bollen
:license: BSD-3-clause
"""
import argparse
import pathlib
from typing import NamedTuple, List
import bitstruct


class ParsedFastqRead(NamedTuple):
    read_id: str
    seq: List[int]
    qualities: List[int]

    def serialize(self) -> 'RawFastqRead':
        pass

    def compress(self) -> bytearray:
        pass

    @classmethod
    def from_bytearray(cls, bytearray) -> 'ParsedFastqRead':
        pass


class RawFastqRead(NamedTuple):
    read_id: str
    seq: List[str]
    qualities: List[str]

    def deserialize(self) -> ParsedFastqRead:
        pass


class RawRastqReader(object):

    def __init__(self, path: pathlib.Path):
        self.path = path
        self.handle = self.path.open("r")
        self.bucket = []

    def __next__(self) -> RawFastqRead:
        i = 0
        while i < 3:
            line = next(self.handle)
            if line == "+\n":
                continue
            self.bucket.append(line)
            i += 1
        read = RawFastqRead(
            read_id=self.bucket[0].strip(),
            seq=list(self.bucket[1].strip()),
            qualities=list(self.bucket[2].strip())
        )
        self.bucket = []
        return read

    def __iter__(self):
        return self

    def close(self):
        self.handle.close()


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
