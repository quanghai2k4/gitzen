// Package limits định nghĩa các hằng số giới hạn để tối ưu hiệu năng
// khi làm việc với các repository lớn.
package limits

const (
	// MaxCommits là số lượng commit tối đa được tải vào danh sách Commits pane.
	// Tăng giá trị này có thể làm chậm quá trình khởi động với repo lớn.
	MaxCommits = 200

	// MaxReflogEntries là số lượng reflog entry tối đa được tải.
	MaxReflogEntries = 100

	// MaxDiffLines là số dòng diff tối đa hiển thị trong diff view.
	// Giới hạn này tránh render quá nhiều dữ liệu cho các file lớn.
	MaxDiffLines = 5000

	// MaxStashEntries là số lượng stash entry tối đa được tải.
	MaxStashEntries = 50

	// CmdTimeout là timeout mặc định (giây) cho các lệnh git thông thường.
	CmdTimeoutSec = 3

	// DiffTimeoutSec là timeout (giây) cho các lệnh tạo diff (lớn hơn do diff lớn).
	DiffTimeoutSec = 10

	// NetworkTimeoutSec là timeout (giây) cho các lệnh cần kết nối mạng
	// như push, pull, fetch.
	NetworkTimeoutSec = 30
)
