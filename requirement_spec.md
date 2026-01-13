# GitZen – Developer Requirements (Linux-first)

## 0. Summary

GitZen là một **TUI Git client** trong terminal (inspired by lazygit), viết bằng **Go + Bubble Tea**. App hiển thị **File Status / Commit Log / Diff Preview** và cho phép **stage/unstage/commit** bằng phím tắt. Sản phẩm phải đóng gói được dưới nhiều định dạng: **raw binary, tar.gz, .deb**.

---

## 1. Goals

### Must-have goals

* Chạy ổn trên **Linux** (Ubuntu/Debian target)
* UI TUI 3 panes, điều hướng bằng keyboard
* Tích hợp Git bằng cách **gọi `git` CLI** (không dùng libgit2/go-git trong MVP)
* Packaging: build binary + tar.gz + .deb
* Demo được manual vs packaged release

### Nice-to-have goals

* Filter/search list
* Branch list + checkout
* Stash
* Config file

---

## 2. Non-goals (Out of scope)

* GUI (windowed)
* Rebase/cherry-pick/merge conflict resolver
* CI/CD automation bắt buộc
* Hỗ trợ Windows/macOS hoàn chỉnh (chỉ thiết kế mở rộng)

---

## 3. Target environment

* OS: Linux (Ubuntu 20.04+ / Debian 11+)
* Terminal: xterm-compatible, UTF-8
* Dependency runtime: `git` (>= 2.x)
* Go toolchain: Go 1.21+ (dev/build)

---

## 4. User stories

1. Là dev, tôi mở GitZen trong repo để xem file changed + staged/unstaged nhanh.
2. Tôi muốn stage/unstage file bằng 1 phím.
3. Tôi muốn xem diff của file đang chọn.
4. Tôi muốn xem commit log gần đây.
5. Tôi muốn commit với message nhập trong TUI.

---

## 5. CLI entry & repo detection

### 5.1 Invocation

* Command: `gitzen` (binary name)
* Optional flags (MVP):

  * `--repo <path>`: mở repo tại path
  * `--version`: in version
  * `--help`: help

### 5.2 Repo detection rules

* Nếu `--repo` có: dùng path đó
* Nếu không: dùng current working directory
* Validate repo bằng:

  * `git rev-parse --show-toplevel`
* Nếu không phải repo:

  * Hiển thị màn hình lỗi trong TUI (hoặc exit với message rõ ràng)
  * Exit code = `2`

**Acceptance criteria**

* Chạy `gitzen` trong repo → vào UI
* Chạy ngoài repo → báo lỗi “Not a git repository” + hướng dẫn

---

## 6. UI Requirements (TUI)

### 6.1 Layout

Tối thiểu 3 panes:

1. **Files Pane (Left, ~35% width)**

   * Tabs/sections: `Unstaged`, `Staged`
2. **Commits Pane (Top-right, ~65% width, ~50% height)**
3. **Diff Pane (Bottom-right, ~65% width, ~50% height)**

Responsive:

* Khi terminal nhỏ hơn (width < 90 hoặc height < 24):

  * Cho phép fallback layout: chỉ 2 panes (Files + Diff) hoặc show warning “terminal too small”

### 6.2 Focus model

* Focus có thể ở: `Files`, `Commits`, `Diff` (diff chỉ đọc)
* Hiển thị focus bằng style (border/title đổi màu hoặc highlight)

### 6.3 Rendering constraints

* Không flicker quá nhiều
* Diff pane hỗ trợ scroll (ít nhất up/down)
* List panes hỗ trợ scroll khi dài

**Acceptance criteria**

* Resize terminal → layout cập nhật
* Focus chuyển pane hoạt động ổn, không crash

---

## 7. Keybindings (MVP)

Global:

* `q` → quit
* `tab` → next pane focus
* `shift+tab` (optional) → previous pane

Navigation:

* `j/k` → down/up (trong pane đang focus)
* `g/G` (optional) → top/bottom

Actions (khi focus Files pane):

* `s` → stage file đang chọn (nếu ở Unstaged)
* `u` → unstage file đang chọn (nếu ở Staged)
* `enter` → toggle xem diff của file (hoặc chỉ cần tự động update diff theo selection)

Actions (khi focus Commits pane):

* `enter` → diff commit đang chọn (optional; MVP chỉ diff working tree cũng OK)

Commit flow:

* `c` → mở input box “Commit message”

  * `enter` → confirm commit
  * `esc` → cancel

**Acceptance criteria**

* Stage/unstage chạy đúng lệnh git, UI refresh sau thao tác
* Commit tạo commit thật (nếu có staged changes), báo lỗi nếu không có gì để commit

---

## 8. Data sources & Git commands (MVP)

### 8.1 Status files

* Command:

  * `git status --porcelain=v1 -z` (khuyên dùng `-z` để parse chắc)
* Parse output thành 2 lists:

  * `Unstaged`: những file có change chưa staged
  * `Staged`: những file đã staged

**Rules parse (porcelain v1)**

* 2 ký tự status: `XY`

  * `X` = index status (staged)
  * `Y` = working tree status (unstaged)
* File thuộc `Staged` nếu `X != ' '`
* File thuộc `Unstaged` nếu `Y != ' '`

File display:

* Hiển thị `path` + status code (ví dụ `M`, `A`, `D`, `??`)
* Với rename/copy: hiển thị `old -> new` (nice-to-have)

### 8.2 Commit log

* Command:

  * `git log --oneline --decorate -n 200`
* Parse:

  * `hash` (short)
  * `message` (rest of line)
  * `decorate` có thể giữ nguyên string

### 8.3 Diff view

Theo selection:

* Nếu chọn file Unstaged:

  * `git diff -- <file>`
* Nếu chọn file Staged:

  * `git diff --staged -- <file>`
* (Optional) Nếu focus Commits:

  * `git show <hash> --stat` hoặc `git show <hash>`

### 8.4 Stage/Unstage

* Stage:

  * `git add -- <file>`
* Unstage:

  * `git restore --staged -- <file>`

### 8.5 Commit

* `git commit -m "<msg>"`
* Nếu commit fail:

  * show error modal/toast

**Acceptance criteria**

* Parse status ổn với file tên có space (nhờ `-z`)
* Diff hiển thị đúng theo staged/unstaged

---

## 9. State management (Bubble Tea Model)

### 9.1 Core state

* `repoRoot string`
* `focus PaneType` (Files/Commits/Diff)
* `filesUnstaged []FileItem`
* `filesStaged []FileItem`
* `commits []CommitItem`
* `selectedUnstaged int`
* `selectedStaged int`
* `selectedCommit int`
* `activeFileList` enum: Unstaged/Staged
* `diffText string`
* `diffScroll int`
* `statusMessage string` (toast)
* `errorMessage string` (modal)

### 9.2 Update triggers

* On start: load status + commits + diff
* After any action (stage/unstage/commit): refresh status + diff (+ commits nếu commit)
* On selection change: refresh diff (debounce optional)

---

## 10. Error handling requirements

### Git not installed

* Detect bằng `exec.LookPath("git")`
* Show error + exit code 3

### Git command failure

* Capture stderr
* Show modal “Git error: …”
* Không crash, giữ UI sống

### Permission / path issues

* Nếu file path invalid hoặc repo inaccessible → show error + exit code phù hợp

---

## 11. Performance requirements

* Status refresh < 300ms trên repo vừa (<= 5k files changed)
* Diff load có thể chậm nhưng không block UI:

  * chạy git diff trong goroutine + update model khi xong (Bubble Tea cmd)

---

## 12. Packaging & release requirements

### 12.1 Versioning

* `VERSION` embed qua ldflags:

  * `-ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD)"`

### 12.2 Output artifacts

* `gitzen-linux-amd64` (binary)
* `gitzen-linux-amd64.tar.gz` (bundle)
* `gitzen_1.0.0_amd64.deb`

### 12.3 tar.gz contents

```
gitzen/
  gitzen
  README.md
  LICENSE
```

### 12.4 .deb requirements

* Install path: `/usr/bin/gitzen`
* Package metadata:

  * Name: `gitzen`
  * Maintainer: <team>
  * Depends: `git (>= 2.0)`
* Post-install message optional

**Acceptance criteria**

* User cài `.deb` xong chạy `gitzen` ở bất kỳ thư mục nào
* Uninstall sạch (nếu có)

---

## 13. Deliverables

* Source code repo (Go)
* README: build/run/packaging instructions
* 3 artifacts (binary, tar.gz, deb)
* Demo script (các command build + packaging + install)

---

## 14. Definition of Done (DoD)

MVP được coi là xong khi:

* TUI 3 panes hoạt động
* Stage/unstage/commit chạy đúng
* Diff preview theo selection
* Packaging tạo được tar.gz và .deb cài được trên Ubuntu/Debian
* Có README + demo script tái lập được

---
