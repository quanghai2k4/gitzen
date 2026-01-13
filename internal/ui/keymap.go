package ui

// Binding định nghĩa một keybinding
type Binding struct {
	Keys   []string // Keys để trigger (e.g., ["space", "enter"])
	Help   string   // Text hiển thị trong info bar
	Action string   // Tên action nội bộ
}

// Keymap chứa tất cả keybindings theo context
type Keymap struct {
	Global   []Binding
	Files    []Binding
	Branches []Binding
	Commits  []Binding
	Stash    []Binding
	Main     []Binding
	CmdLog   []Binding
}

// DefaultKeymap - keybindings mặc định (lazygit-style)
var DefaultKeymap = Keymap{
	Global: []Binding{
		{Keys: []string{"tab"}, Help: "next pane", Action: "focus_next"},
		{Keys: []string{"shift+tab"}, Help: "prev pane", Action: "focus_prev"},
		{Keys: []string{"q"}, Help: "quit", Action: "quit"},
		{Keys: []string{"p"}, Help: "pull", Action: "git_pull"},
		{Keys: []string{"P"}, Help: "push", Action: "git_push"},
		{Keys: []string{"f"}, Help: "fetch", Action: "git_fetch"},
		{Keys: []string{"R"}, Help: "refresh", Action: "refresh_all"},
		{Keys: []string{"1"}, Help: "files", Action: "focus_files"},
		{Keys: []string{"2"}, Help: "branches", Action: "focus_branches"},
		{Keys: []string{"3"}, Help: "commits", Action: "focus_commits"},
		{Keys: []string{"4"}, Help: "stash", Action: "focus_stash"},
		{Keys: []string{"5"}, Help: "main", Action: "focus_main"},
	},
	Files: []Binding{
		{Keys: []string{"space"}, Help: "stage/unstage", Action: "toggle_stage"},
		{Keys: []string{"a"}, Help: "stage all", Action: "stage_all"},
		{Keys: []string{"c"}, Help: "commit", Action: "open_commit"},
		{Keys: []string{"A"}, Help: "amend", Action: "open_amend"},
		{Keys: []string{"d"}, Help: "discard", Action: "discard_changes"},
		{Keys: []string{"e"}, Help: "edit", Action: "edit_file"},
		{Keys: []string{"o"}, Help: "open", Action: "open_file"},
		{Keys: []string{"s"}, Help: "stash", Action: "stash_changes"},
		{Keys: []string{"enter"}, Help: "view diff", Action: "view_file_diff"},
	},
	Branches: []Binding{
		{Keys: []string{"space"}, Help: "checkout", Action: "checkout_branch"},
		{Keys: []string{"n"}, Help: "new branch", Action: "create_branch"},
		{Keys: []string{"d"}, Help: "delete", Action: "delete_branch"},
		{Keys: []string{"D"}, Help: "force delete", Action: "force_delete_branch"},
		{Keys: []string{"r"}, Help: "rebase", Action: "rebase_branch"},
		{Keys: []string{"m"}, Help: "merge", Action: "merge_branch"},
		{Keys: []string{"enter"}, Help: "view commits", Action: "view_branch_commits"},
	},
	Commits: []Binding{
		{Keys: []string{"enter"}, Help: "view", Action: "view_commit"},
		{Keys: []string{"r"}, Help: "revert", Action: "revert_commit"},
		{Keys: []string{"R"}, Help: "reset", Action: "reset_to_commit"},
		{Keys: []string{"c"}, Help: "cherry-pick", Action: "cherry_pick"},
		{Keys: []string{"space"}, Help: "checkout", Action: "checkout_commit"},
	},
	Stash: []Binding{
		{Keys: []string{"space"}, Help: "apply", Action: "stash_apply"},
		{Keys: []string{"p"}, Help: "pop", Action: "stash_pop"},
		{Keys: []string{"d"}, Help: "drop", Action: "stash_drop"},
		{Keys: []string{"enter"}, Help: "view", Action: "view_stash"},
	},
	Main: []Binding{
		{Keys: []string{"j"}, Help: "down", Action: "scroll_down"},
		{Keys: []string{"k"}, Help: "up", Action: "scroll_up"},
		{Keys: []string{"d"}, Help: "page down", Action: "page_down"},
		{Keys: []string{"u"}, Help: "page up", Action: "page_up"},
		{Keys: []string{"g"}, Help: "top", Action: "scroll_top"},
		{Keys: []string{"G"}, Help: "bottom", Action: "scroll_bottom"},
	},
	CmdLog: []Binding{
		{Keys: []string{"j"}, Help: "down", Action: "scroll_down"},
		{Keys: []string{"k"}, Help: "up", Action: "scroll_up"},
		{Keys: []string{"g"}, Help: "top", Action: "scroll_top"},
		{Keys: []string{"G"}, Help: "bottom", Action: "scroll_bottom"},
	},
}

// InfoBarHelp trả về help text cho info bar dựa trên focused pane
func (k Keymap) InfoBarHelp(focusedPane string) string {
	var bindings []Binding

	switch focusedPane {
	case "files":
		bindings = k.Files
	case "branches":
		bindings = k.Branches
	case "commits":
		bindings = k.Commits
	case "stash":
		bindings = k.Stash
	case "main":
		bindings = k.Main
	case "cmdlog":
		bindings = k.CmdLog
	default:
		bindings = k.Global
	}

	// Format: key: help | key: help | ...
	result := ""
	for i, b := range bindings {
		if i > 0 {
			result += " | "
		}
		key := b.Keys[0]
		result += key + ": " + b.Help
	}

	return result
}
