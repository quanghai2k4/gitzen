package git

import "bytes"

type FileItem struct {
	Path   string
	Status string
	Staged bool
}

type Status struct {
	Staged   []FileItem
	Unstaged []FileItem
}

func ParseStatusPorcelainV1Z(data []byte) Status {
	var staged []FileItem
	var unstaged []FileItem
	for _, entry := range bytes.Split(data, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		if len(entry) < 3 {
			continue
		}

		x := entry[0]
		y := entry[1]
		path := string(bytes.TrimSpace(entry[2:]))
		if path == "" {
			continue
		}

		if x != ' ' {
			staged = append(staged, FileItem{Path: path, Status: string(x), Staged: true})
		}
		if y != ' ' {
			unstaged = append(unstaged, FileItem{Path: path, Status: string(y), Staged: false})
		}
	}
	return Status{Staged: staged, Unstaged: unstaged}
}
