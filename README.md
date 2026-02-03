# Docker Images CSV Exporter

Tool để xuất danh sách tất cả Docker images ra file CSV với đầy đủ các cột thông tin.

## Yêu cầu

- Go 1.16 trở lên
- Docker daemon đang chạy

## Cài đặt

```bash
go mod tidy
go build
```

## Sử dụng

```bash
./docker-images-csv
```

Chương trình sẽ tạo file `docker_images.csv` chứa thông tin tất cả Docker images.

## Các cột trong CSV

- **ID**: Docker image ID (12 ký tự)
- **Repository**: Tên repository
- **Tag**: Tag của image
- **Created**: Timestamp tạo image
- **Size (MB)**: Kích thước image (MB)
- **SharedSize (MB)**: Kích thước chia sẻ (MB)
- **VirtualSize (MB)**: Kích thước ảo (MB)
- **Containers**: Số lượng container đang dùng image
- **Labels**: Labels của image (định dạng key=value)

## Ví dụ output

```csv
ID,Repository,Tag,Created,Size (MB),SharedSize (MB),VirtualSize (MB),Containers,Labels
a1b2c3d4e5f6,nginx,latest,1706950400,142.50,0.00,142.50,0,maintainer=NGINX Docker Maintainers
```

## Lưu ý

- Nếu một image có nhiều tags, sẽ có nhiều dòng tương ứng
- Images không có tag sẽ hiển thị `<none>`
- Cần quyền truy cập Docker daemon
