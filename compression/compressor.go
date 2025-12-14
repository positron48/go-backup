package compression

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compressor interface {
	Compress(source, destination string) error
}

type GzipCompressor struct{}

func (c *GzipCompressor) Compress(source, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	writer := gzip.NewWriter(dstFile)
	defer writer.Close()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	return nil
}

type ZipCompressor struct{}

func (c *ZipCompressor) Compress(source, destination string) error {
	zipFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	// Если source - это файл
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if !info.IsDir() {
		return c.addFileToZip(writer, source, filepath.Base(source))
	}

	// Если source - это директория
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Пропускаем файлы/директории, к которым нет доступа
			return nil
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Пропускаем специальные файлы (socket, named pipe, device files)
		mode := info.Mode()
		if mode&os.ModeSocket != 0 || mode&os.ModeNamedPipe != 0 || mode&os.ModeDevice != 0 {
			return nil
		}

		// Проверяем, является ли это симлинком, указывающим на директорию
		if mode&os.ModeSymlink != 0 {
			// Проверяем, куда указывает симлинк
			target, err := os.Readlink(path)
			if err != nil {
				// Не удалось прочитать симлинк, пропускаем
				return nil
			}
			// Получаем абсолютный путь цели
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			// Проверяем, находится ли цель внутри исходной директории
			relTarget, err := filepath.Rel(source, target)
			if err != nil || strings.HasPrefix(relTarget, "..") {
				// Цель находится вне исходной директории, пропускаем
				return nil
			}
			// Проверяем, является ли цель директорией
			// Если цель не существует, просто пропускаем (не добавляем битый симлинк)
			targetInfo, err := os.Stat(target)
			if err != nil {
				// Цель не существует, пропускаем
				return nil
			}
			if targetInfo.IsDir() {
				// Симлинк указывает на директорию, пропускаем
				return nil
			}
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		return c.addFileToZip(writer, path, relPath)
	})
}

func (c *ZipCompressor) addFileToZip(writer *zip.Writer, filePath, zipPath string) error {
	// Используем Lstat, чтобы не следовать симлинкам
	info, err := os.Lstat(filePath)
	if err != nil {
		// Файл не существует или недоступен, пропускаем
		return nil
	}

	// Дополнительная проверка: если это директория, пропускаем
	if info.IsDir() {
		return nil
	}

	// Пропускаем специальные файлы (socket, named pipe, device files)
	mode := info.Mode()
	if mode&os.ModeSocket != 0 || mode&os.ModeNamedPipe != 0 || mode&os.ModeDevice != 0 {
		return nil
	}

	// Если это симлинк, проверяем, что цель существует и это файл
	if mode&os.ModeSymlink != 0 {
		target, err := os.Readlink(filePath)
		if err != nil {
			// Не удалось прочитать симлинк, пропускаем
			return nil
		}
		// Получаем абсолютный путь цели
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(filePath), target)
		}
		// Проверяем, что цель существует и это файл
		// Если цель не существует, просто добавляем симлинк как есть (без разыменования)
		targetInfo, err := os.Stat(target)
		if err != nil {
			// Цель не существует, добавляем симлинк как есть
			// Не меняем info, используем информацию о самом симлинке
		} else if targetInfo.IsDir() {
			// Цель - директория, пропускаем
			return nil
		} else {
			// Используем информацию о цели для создания заголовка
			info = targetInfo
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		// Если не удалось открыть файл, пропускаем
		return nil
	}
	defer file.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = zipPath
	header.Method = zip.Deflate

	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

type TarCompressor struct{}

func (c *TarCompressor) Compress(source, destination string) error {
	tarFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	writer := tar.NewWriter(tarFile)
	defer writer.Close()

	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if !info.IsDir() {
		return c.addFileToTar(writer, source, filepath.Base(source))
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		return c.addFileToTar(writer, path, relPath)
	})
}

func (c *TarCompressor) addFileToTar(writer *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	header.Name = tarPath

	if err := writer.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

type TarGzCompressor struct{}

func (c *TarGzCompressor) Compress(source, destination string) error {
	// Сначала создаем tar во временный файл
	tmpTar := destination + ".tmp.tar"
	if err := (&TarCompressor{}).Compress(source, tmpTar); err != nil {
		return err
	}
	defer os.Remove(tmpTar)

	// Затем сжимаем gzip
	tarFile, err := os.Open(tmpTar)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer tarFile.Close()

	gzFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer gzFile.Close()

	writer := gzip.NewWriter(gzFile)
	defer writer.Close()

	_, err = io.Copy(writer, tarFile)
	if err != nil {
		return fmt.Errorf("failed to compress tar: %w", err)
	}

	return nil
}

type NoCompressor struct{}

func (c *NoCompressor) Compress(source, destination string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func NewCompressor(compressionType string) (Compressor, error) {
	switch strings.ToLower(compressionType) {
	case "gzip":
		return &GzipCompressor{}, nil
	case "zip":
		return &ZipCompressor{}, nil
	case "tar":
		return &TarCompressor{}, nil
	case "tar.gz":
		return &TarGzCompressor{}, nil
	case "none", "":
		return &NoCompressor{}, nil
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
}

