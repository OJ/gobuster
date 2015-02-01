Gobuster v0.2 (OJ Reeves @TheColonial)
======================================

Alternative directory and file busting tool written in Go.

### Oh dear God.. WHY!?

Because I wanted:

1. ... something that didn't have a fat Java GUI (console FTW).
1. ... to build something that just worked on the command line.
1. ... something that did not do recursive brute force.
1. ... something that allowed me to brute for folders and multiple extensions at once.
1. ... something that compiled to native on multiple platforms.
1. ... something that was faster than an interpreted script (such as Python).
1. ... something that didn't require a runtime.
1. ... use something that was good with concurrency (hence Go).
1. ... to build something in Go that wasn't totally useless.

### But it's shit! And your implementation sucks!

Yes, you're probably correct. Feel free to :

* Not use it.
* Show me how to do it better.

### Command line options

* `-c=<http cookies>` - use this to specify any cookies that you might need (simulating auth).
* `-f=<true|false>` - set to `true` if you want to append `/` for directory brute forces.
* `-s=<status codes>` - comma-separated set of the list of status codes to be deemed a "positive" (default: `200,204,301,302,307`).
* `-t=<threads>` - number of threads to run (default: `10`).
* `-u=<url>` - full to the folder to brute force, including scheme.
* `-v=<true|false>` - verbose output.
* `-w=<wordlist>` - path to the wordlist used for brute forcing.
* `-x=<extensions>` - list of extensions to check for, if any.

### Examples
```
$ ./gobuster -u=https://mysite.com/path/to/folder '-c=session=123456' -t=50 -w=common-files.txt -x=.php,.html
```

### License

See the LICENSE file.
