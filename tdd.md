--

# TÀI LIỆU THIẾT KẾ UI: GOLANG GIT TUI

## 1. Tổng quan Kiến trúc (Architecture Overview)

Chúng ta sẽ sử dụng kiến trúc **The Elm Architecture (TEA)** thông qua thư viện `Bubble Tea`. Đây là mô hình chuẩn mực cho các ứng dụng TUI hiện đại bằng Go, giúp tách biệt rõ ràng giữa **State** (Dữ liệu), **View** (Giao diện) và **Update** (Logic).

### Sơ đồ luồng dữ liệu:

1. **Model:** Chứa toàn bộ trạng thái của UI (vị trí con trỏ, nội dung file, pane đang active...).
2. **View:** Hàm thuần túy (pure function) nhận `Model` và trả về `string` để render ra terminal.
3. **Update:** Nhận `Msg` (sự kiện phím bấm, dữ liệu từ git cmd) và trả về `Model` mới.
4. **Cmd:** Các tác vụ bất đồng bộ (chạy lệnh git, đọc file IO) trả về `Msg` để feed lại vào vòng lặp.

### Tech Stack đề xuất:

* **Core Framework:** `github.com/charmbracelet/bubbletea` (Quản lý vòng đời ứng dụng).
* **Styling & Layout:** `github.com/charmbracelet/lipgloss` (CSS cho Terminal: màu sắc, border, padding).
* **Git Interface:** `os/exec` (gọi lệnh git native) hoặc `github.com/go-git/go-git` (thư viện native Go). *Khuyên dùng `os/exec` để tương thích tốt nhất với config git của user.*

---

## 2. Thiết kế Bố cục (Layout Strategy)

Giao diện sẽ chia thành hệ thống lưới (Grid) linh hoạt, tự động điều chỉnh khi thay đổi kích thước cửa sổ (Window Resize).

### 2.1. Phân chia màn hình (Panes)

Màn hình được chia làm 3 khu vực chính:

| Khu vực | Tên Component | Vị trí | Chức năng |
| --- | --- | --- | --- |
| **Zone A** | `SideBar` | Bên trái (30% width) | Chứa các danh sách điều hướng (Status, Files, Branches, Commits). Được xếp chồng dọc (Vertical Stack). |
| **Zone B** | `MainView` | Bên phải (70% width) | Hiển thị nội dung chi tiết: Diff của file, nội dung file, hoặc log đồ thị. |
| **Zone C** | `CommandLog` | Dưới cùng (Bottom) | Hiển thị lệnh Git đang chạy hoặc thanh trạng thái/phím tắt hướng dẫn. |

### 2.2. Modal & Popup

* Sử dụng cơ chế **Layering**: Vẽ một hộp (Box) đè lên `MainView` khi cần nhập liệu (ví dụ: Commit Message, Rename Branch) hoặc hiển thị Help Menu (`?`).

---

## 3. Cấu trúc Dữ liệu (Go Data Structures)

Đây là phần cốt lõi để hiện thực hóa thiết kế.

### 3.1. Global Model

```go
type Model struct {
    // Kích thước Terminal hiện tại (để tính toán layout)
    Width  int
    Height int

    // Quản lý Focus (Tiêu điểm)
    ActivePane PaneID // Enum: PaneFiles, PaneBranches, PaneCommits, PaneMain...
    
    // Trạng thái các Components con
    StatusBox   StatusModel
    FileBox     ListModel
    BranchBox   ListModel
    CommitBox   ListModel
    MainView    ViewModel
    
    // Trạng thái ứng dụng
    IsLoading   bool // Hiển thị spinner nếu đang chạy lệnh git nặng
    ErrorMsg    string
}

```

### 3.2. List Component (Tái sử dụng cho Files, Branches, Commits)

Để code sạch, chúng ta tạo một struct `ListModel` chung cho các danh sách bên trái.

```go
type ListModel struct {
    Title        string
    Items        []ListItem // Interface chứa Title, Description
    Cursor       int        // Vị trí con trỏ hiện tại
    Selected     map[int]bool // Dùng cho multi-select (ví dụ: stage nhiều file)
    IsFocused    bool       // Có đang được active không? (Để đổi màu border)
}

```

---

## 4. Thiết kế Tương tác (Interaction Design)

### 4.1. Keybindings (Phím tắt)

Hệ thống phím tắt chia làm 2 tầng: **Global** (luôn hoạt động) và **Contextual** (phụ thuộc Pane đang focus).

* **Global Navigation:**
* `Tab` / `Shift+Tab`: Chuyển đổi vòng tròn giữa các Panes (Files -> Branches -> Commits -> Main).
* `Ctrl+c` / `q`: Thoát ứng dụng.


* **List Navigation (Vim style):**
* `j` / `Down`: Xuống dòng.
* `k` / `Up`: Lên dòng.
* `PageUp` / `PageDown`: Cuộn nhanh.


* **Actions:**
* `Space`: Toggle trạng thái (Stage/Unstage file).
* `Enter`: Đi vào chi tiết (Focus sang MainView để edit hoặc xem full diff).
* `c`: Mở popup commit.



### 4.2. Focus Styling (Trải nghiệm người dùng)

Để người dùng biết họ đang ở đâu, UI phải phản hồi trực quan:

* **Pane Active:** Border màu sáng (ví dụ: `ActiveBorderColor = lipgloss.Color("62")` - Tím/Xanh).
* **Pane Inactive:** Border màu tối/xám (`InactiveBorderColor = lipgloss.Color("236")`).
* **Cursor:** Dòng đang chọn được đảo màu nền (Reverse video) hoặc có ký tự `>` đằng trước.

---

## 5. Quy trình Render (Render Loop)

Sử dụng `lipgloss` để ghép nối các chuỗi string lại với nhau.

```go
func (m Model) View() string {
    // 1. Tính toán kích thước cho từng pane dựa trên m.Height và m.Width
    lhsWidth := m.Width / 3
    rhsWidth := m.Width - lhsWidth
    
    // 2. Render cột trái (Left Hand Side)
    // Ghép dọc các box: Status, Files, Branches, Commits
    lhs := lipgloss.JoinVertical(lipgloss.Left,
        m.StatusBox.View(lhsWidth),
        m.FileBox.View(lhsWidth),
        m.BranchBox.View(lhsWidth),
        m.CommitBox.View(lhsWidth),
    )

    // 3. Render cột phải (Main View)
    rhs := m.MainView.View(rhsWidth)

    // 4. Ghép ngang 2 cột
    fullView := lipgloss.JoinHorizontal(lipgloss.Top, lhs, rhs)

    // 5. Nếu có Modal (Popup), render đè lên fullView
    if m.ShowCommitModal {
        return overlay(fullView, m.CommitModal.View())
    }

    return fullView
}

```

---

## 6. Kế hoạch Implement (Roadmap)

1. **Phase 1: Skeleton & Layout**
* Dựng struct `Model` cơ bản.
* Sử dụng `lipgloss` vẽ ra layout tĩnh chia 2 cột, cột trái chia 4 dòng.
* Xử lý sự kiện `WindowSizeMsg` để layout co giãn được.


2. **Phase 2: Navigation Logic**
* Implement logic di chuyển con trỏ (`j`, `k`) trong danh sách giả (dummy data).
* Implement logic chuyển focus (`Tab`) đổi màu viền các pane.


3. **Phase 3: Git Integration (Real Data)**
* Viết hàm `getGitStatus()` chạy lệnh `git status -s` và parse kết quả vào struct.
* Viết hàm `getGitLog()` chạy lệnh `git log --oneline`.
* Map dữ liệu thật vào UI.


4. **Phase 4: Actions & Interactive**
* Xử lý phím `Space` -> chạy `git add <file>`.
* Tự động refresh UI sau khi chạy lệnh git.


5. **Phase 5: Polishing**
* Thêm màu sắc cho trạng thái file (Xanh = New, Vàng = Modified, Đỏ = Deleted).
* Thêm Spinner loading animation.



---
