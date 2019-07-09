fastqube
========

NOTE: THIS IS A WORK IN PROGRESS. SEE SECTION "Currently implemented" FOR
CURRENTLY IMPLEMENTED FUNCTIONALITIES.  

Fastqube is little tool that makes a binary representation of fastq file(s).

Fastq files are typically very inefficient text files, and are as such
usually directly compressed with gzip.

Many tools exist that transform fastq into a better compressible format,
but those typically rely on aligning to a reference genome first (e.g CRAM).

Instead, fastqube is a simple direct binary representation of a given fastq
file. As such, each read consists of three fields:

1. A read ID in UTF-8.
2. The sequence encoded in 3-bit encoding (ACTGN, other IUPAC codes
not supported)
3. The quality string encoded in 6-bit encoding. Only sanger-encoded qualities
are supported.

Fastqube files also contain a 4096-byte header, which contains some metadata
about the program and binary mode used, and reserves space for potential future
metadata fields.

# Lossy modes

Apart from the lossless mode described above, fastqube also supports three
different lossy modes through various settable parameters. These parameters
can be combined.

## ID size

Read IDs are fixed-with, but nevertheless are settable. When setting a read ID
size of 0, read IDs are not emitted to the compressed stream at all.
When serializing fastqube files back to fastq files, read IDs are generated as
a simple integer, optionally with a pair tag. I.e. the n-th read will have an
ID of `n`.

## Block Quality encoding

With block quality encoding enabled, quality scores are stored in 3-bit
representation, with only five possible values: 0, 2, 26, 31 and 41. Any other
scores will be rounded down to the nearest possible value.

## 2-Bit Sequence encoding

With 2-bit sequencing encoding enabled, the sequence is stored in a 2-bit
representation. N, and all other IUPAC symbols, will be squashed to a G.
G is also the base the Novaseq uses as its 'black' color, hence we hope that
choosing this base as the fallback mimics the NovaSeq's behavior.


# Currently implemented

* ✔️ Compression in lossless mode
* ✔️ Compression with 2-bit sequences
* ✔️ Compression with block Qualities
* ✔️ Compression with settable bytes per read ID


# Limitations

1. IUPAC codes other than N are not supported. We may offer a feature in the
future that squashes all non-N IUPAC codes to N, but this can naturally
only be supported in a lossy mode.
2. All reads must have the same length. We may offer a feature in the future
that supports variable length reads.
3. Read IDs may not be longer than 512 bits (64 bytes). We may offer a feature
in the future that supports settable read ID lengths.
4. Only sanger-encoded quality strings are supported.
5. Reads longer than 65536 bases are currently not supported


# How much smaller are fastqube files compared to fastq files

A typical fastq read with a 64-byte ID and 150 bases consists of 364 bytes.
The direct lossless binary representation of such a read would consist of
about 234 bytes.

With the lossy modes the size is further trimmed to:

1. About 170 bytes with read ID size of 0.
2. About 178 bytes in `Block Quality Mode`
3. About 114 bytes in `Block Quality Mode` _and_ a read ID size of 0.
4. About 95 bytes in `Block Quality Mode` _and_ a read ID size of 0, _and_
2-bit sequence encoding.


# Usage

Single end modes optionally read from stdin.

## Lossless compression
```bash
fastqube -c input.fastq > output.fqb
```

### Paired-end mode
```bash
fastqube -c -R1 R1.fastq -R2 R2.fastq -o1 R1.fqb -o2 R2.fqb
```

## Lossless decompression
```bash
fastqube -d input.fqb > output.fastq
```

### Paired-end mode
```bash
fastqube -d -R1 R1.fqb -R2 R2.fqb -o1 R1.fastq -o2 R2.fastq
```

## Lossy compression
Lossly compression is enabled with several parameters:

### Read ID size

```bash
fastqube -B 0 -c input.fastq > output.fqb
```

### 2-bit sequence encoding

```bash
fastqube -2 -c input.fastq > output.fqb
```

### block quality encoding

```bash
fastqube -b -c input.fastq > output.fqb
```

### Combining all threeBitDNA

```bash
fastqube -b -B 0 -2 -c input.fastq > output.fqb
```

## Lossy decompression

Lossy decompression uses the same command-line as lossless decompression:
the fastqube header contains the information about which lossy mode was used.


# Prototype

There is a little [prototype](prototype.py) implemented in python. It so far
only supports compression in lossless mode. You can use this prototype
to get an impression of how fastqube-generated files will ultimately look like.

As this is a prototype, it is really slow, and even the file format is liable
to change.    


# License

BSD-3-clause
