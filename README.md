## data-diff

data-diff is tool that can be used to create signature from basis file and delta from changed file.

Resulted signature contains list of chunks which have minimum size of 32 bytes and maximum size of 1024 bytes. Chunks are separated by specific hash numbers (numbers which last 7 bits are 1's) generated with rolling hash algorithm.

data-diff delta file is in format that rdiff tool supports for checking functionality with rdiff's patch command. 
(tested with version librsync 2.0.2)

### Build

```
go build data-diff
```

### Usage

```
Usage: data-diff [OPTIONS] signature [BASIS [SIGNATURE]]
                 [OPTIONS] delta SIGNATURE [NEWFILE [DELTA]]

Options:
-v, --verbose             Trace internal processing
-?, --help                Show this help message
-f, --force               Force overwriting existing files
```