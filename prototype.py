"""
fastqube prototype
~~~~~~~~~~~~~~~~~~

:copyright: (c) 2019 Sander Bollen
:license: BSD-3-clause
"""
import argparse
import datetime
import pathlib
import sys
from typing import NamedTuple, List
import bitstruct


SEQ_MAP = {
    "A": 0,
    "T": 1,
    "C": 2,
    "G": 3,
    "N": 4
}

SEQ_MAP_REV = {v: k for k, v in SEQ_MAP.items()}


# bistruct header format
HEADER_FMT = "t32768"


# bitstruct format for 150-base read
READ_FMT = "t512" + ("u3"*150) + ("u6"*150)


def create_read_fmt(read_id: str) -> str:
    """create a bitstruct format for any given read it up to 512 bits"""
    if len(read_id) > 64:
        raise ValueError("Read ID is too large")
    padding = 512 - len(read_id)*8
    r_fmt = "t{0}p{1}".format(len(read_id)*8, padding)
    return r_fmt + ("u3" * 150) + ("u6" * 150)


class ParsedFastqRead(NamedTuple):
    read_id: str
    seq: List[int]
    qualities: List[int]

    def serialize(self) -> 'RawFastqRead':
        return RawFastqRead(
            read_id=self.read_id,
            seq=[SEQ_MAP_REV[x] for x in self.seq],
            qualities=[chr(i+33) for i in self.qualities]
        )

    def compress(self) -> bytearray:
        fmt = create_read_fmt(self.read_id)
        args = [self.read_id] + self.seq + self.qualities
        return bitstruct.pack(fmt, *args)

    @classmethod
    def from_bytearray(cls, bytearray) -> 'ParsedFastqRead':
        unpacked = bitstruct.unpack(READ_FMT, bytearray)
        read_id = unpacked[0]
        seq = unpacked[1: 151]
        qual = unpacked[151:]
        return cls(read_id, seq, qual)


class RawFastqRead(NamedTuple):
    read_id: str
    seq: List[str]
    qualities: List[str]

    def deserialize(self) -> ParsedFastqRead:
        return ParsedFastqRead(
            read_id=self.read_id,
            seq=[SEQ_MAP[x.upper()] for x in self.seq],
            qualities=[ord(x)-33 for x in self.qualities]
        )


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


def create_header() -> bytearray:
    val = ("program: fastqube \n"
           "version: PROTOTYPE \n"
           "mode: LOSSLESS\n"
           "date: {0} \n").format(datetime.datetime.utcnow().isoformat())
    padding = (4096*8) - (len(val)*8)
    fmt = "t{0}p{1}".format(len(val)*8, padding)
    return bitstruct.pack(fmt, val)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("input", type=pathlib.Path)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("-c", "--compress", action="store_true")
    group.add_argument("-d", "--decompress", action="store_true")

    args = parser.parse_args()

    if args.compress:
        sys.stdout.buffer.write(create_header())
        reader = RawRastqReader(args.input)
        for read in reader:
            compressed = read.deserialize().compress()
            sys.stdout.buffer.write(compressed)
        reader.close()
    elif args.decompress:
        raise NotImplementedError
    else:
        raise NotImplementedError
    sys.stdout.buffer.flush()
