* return specific errors and do not mention command line switches in libgobuster
* no log.Printf and fmt.Printf inside libgobuster
* use smth like `tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)` for outputting options (`GetConfigString`)