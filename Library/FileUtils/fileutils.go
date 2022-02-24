package FileUtils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetCurrentAbPathByExecutable 当前执行文件目录
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}
// GetCurrentFuncNameByCaller 当前方法执行函数名
func GetCurrentFuncNameByCaller() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

// GetCurrentAbPath 项目根目录绝对路径
func GetCurrentAbPath() string {
	dir := GetCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		wd, _ := os.Getwd()
		return wd
	}
	if strings.Contains(dir, "/tmp") {
		return filepath.Dir(dir)
	}

	return dir
}

// RealPath 基于构件执行文件的绝对文件路径
func RealPath(fp string) (string, error) {
	if path.IsAbs(fp) {
		return fp, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, fp), err
}



// SelfPath 完整的执行文件绝对路径
func GetCurrPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
// GetCurrPathDir 执行文件目录完整路径
func GetCurrPathDir() string {
	filepathdata,err:=GetCurrPath() 
	if err  == nil {
		return filepath.Dir(filepathdata)	
	}
	return ""
	
}

// GetBaseFilename 从路径中提取文件名
func GetBaseFilename(fp string) string {
	return path.Base(fp)
}

// CreateDirectory 新建不存在的文件夹
func CreateDirectory(dirPath string) error {
	f, e := os.Stat(dirPath)
	if e != nil && os.IsNotExist(e) {
		return os.MkdirAll(dirPath, 0755)
	}
	if e == nil && !f.IsDir() {
		return fmt.Errorf("create dir:%s error, not a directory", dirPath)
	}
	return e
}

// IsExist 检测文件或者目录是否存在
// 不存在的时候将返回 fasle
func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// IsFile checks whether the path is a file,
// it returns false when it's a directory or does not exist.
func IsFile(fp string) bool {
	f, e := os.Stat(fp)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func Remove(filename string) error {
	if IsFile(filename) && IsExist(filename) {
		return os.Remove(filename)
	}
	return nil
}

func ReadBytes(cpath string) ([]byte, error) {
	if !IsExist(cpath) {
		return nil, fmt.Errorf("%s not exists", cpath)
	}

	if !IsFile(cpath) {
		return nil, fmt.Errorf("%s not file", cpath)
	}

	return ioutil.ReadFile(cpath)
}

func ReadString(cpath string) (string, error) {
	bs, err := ReadBytes(cpath)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func ReadStringTrim(cpath string) (string, error) {
	out, err := ReadString(cpath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func ReadJson(cpath string, cptr interface{}) error {
	os.MkdirAll(path.Dir(cpath), os.ModePerm)
	bs, err := ReadBytes(cpath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %s", cpath, err.Error())
	}

	err = json.Unmarshal(bs, cptr)
	if err != nil {
		return fmt.Errorf("cannot parse %s: %s", cpath, err.Error())
	}

	return nil
}

func WriteBytes(filePath string, b []byte) (int, error) {
	os.MkdirAll(path.Dir(filePath), os.ModePerm)
	fw, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer fw.Close()
	return fw.Write(b)
}

func WriteString(filePath string, s string) (int, error) {
	return WriteBytes(filePath, []byte(s))
}

func OpenLogFile(fp string) (*os.File, error) {
	os.MkdirAll(path.Dir(fp), os.ModePerm)
	return os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

// 添加文本
func AppendFile(filePath string, b []byte) error {
	os.MkdirAll(path.Dir(filePath), os.ModePerm)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(string(b) + "\r\n\r\n")

	return nil
}

// list dirs under dirPath
func DirsUnder(dirPath string) ([]string, error) {
	if !IsExist(dirPath) {
		return []string{}, nil
	}

	fs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return []string{}, err
	}

	sz := len(fs)
	if sz == 0 {
		return []string{}, nil
	}

	ret := make([]string, 0, sz)
	for i := 0; i < sz; i++ {
		if fs[i].IsDir() {
			name := fs[i].Name()
			if name != "." && name != ".." {
				ret = append(ret, name)
			}
		}
	}

	return ret, nil
}

// list files under dirPath
func FilesUnder(dirPath string) ([]string, error) {
	if !IsExist(dirPath) {
		return []string{}, nil
	}

	fs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return []string{}, err
	}

	sz := len(fs)
	if sz == 0 {
		return []string{}, nil
	}

	ret := make([]string, 0, sz)
	for i := 0; i < sz; i++ {
		if !fs[i].IsDir() {
			ret = append(ret, fs[i].Name())
		}
	}

	return ret, nil
}
// BufferSize defines the buffer size when reading and writing file.
const BufferSize = 8 * 1024 * 1024


// DeleteFile deletes a file not a directory.
func DeleteFile(filePath string) error {
	if !PathExist(filePath) {
		return fmt.Errorf("delete file:%s error, file not exist", filePath)
	}
	if IsDir(filePath) {
		return fmt.Errorf("delete file:%s error, is a directory instead of a file", filePath)
	}
	return os.Remove(filePath)
}

// DeleteFiles deletes all the given files.
func DeleteFiles(filePaths ...string) {
	if len(filePaths) > 0 {
		for _, f := range filePaths {
			DeleteFile(f)
		}
	}
}

// OpenFile opens a file. If the parent directory of the file isn't exist,
// it will create the directory.
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	if PathExist(path) {
		return os.OpenFile(path, flag, perm)
	}
	if err := CreateDirectory(filepath.Dir(path)); err != nil {
		return nil, err
	}

	return os.OpenFile(path, flag, perm)
}

// Link creates a hard link pointing to src named linkName for a file.
func Link(src string, linkName string) error {
	if PathExist(linkName) {
		if IsDir(linkName) {
			return fmt.Errorf("link %s to %s: error, link name already exists and is a directory", linkName, src)
		}
		if err := DeleteFile(linkName); err != nil {
			return err
		}

	}
	return os.Link(src, linkName)
}

// SymbolicLink creates target as a symbolic link to src.
func SymbolicLink(src string, target string) error {
	// TODO Add verifications.
	return os.Symlink(src, target)
}

// CopyFile copies the file src to dst.
func CopyFile(src string, dst string) (err error) {
	var (
		s *os.File
		d *os.File
	)
	if !IsRegularFile(src) {
		return fmt.Errorf("copy file:%s error, is not a regular file", src)
	}
	if s, err = os.Open(src); err != nil {
		return err
	}
	defer s.Close()

	if PathExist(dst) {
		return fmt.Errorf("copy file:%s error, dst file already exists", dst)
	}

	if d, err = OpenFile(dst, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return err
	}
	defer d.Close()

	buf := make([]byte, BufferSize)
	for {
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 || err == io.EOF {
			break
		}
		if _, err := d.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// MoveFile moves the file src to dst.
func MoveFile(src string, dst string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("move file:%s error, is not a regular file", src)
	}
	if PathExist(dst) && !IsDir(dst) {
		if err := DeleteFile(dst); err != nil {
			return err
		}
	}
	return os.Rename(src, dst)
}

// MoveFileAfterCheckMd5 will check whether the file's md5 is equals to the param md5
// before move the file src to dst.
func MoveFileAfterCheckMd5(src string, dst string, md5 string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("move file with md5 check:%s error, is not a "+
			"regular file", src)
	}
	m := Md5SumFile(src)
	if m != md5 {
		return fmt.Errorf("move file with md5 check:%s error, md5 of source "+
			"file doesn't match against the given md5 value", src)
	}
	return MoveFile(src, dst)
}

// PathExist reports whether the path is exist.
// Any error get from os.Stat, it will return false.
func PathExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// IsDir reports whether the path is a directory.
func IsDir(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.IsDir()
}
func ReadFileContent(path string)( string,error) {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "",err
	}
	return string(content),err

}

// IsRegularFile reports whether the file is a regular file.
// If the given file is a symbol link, it will follow the link.
func IsRegularFile(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.Mode().IsRegular()
}

// Md5Sum generates md5 for a given file.
func Md5SumFile(name string) string {
	if !IsRegularFile(name) {
		return ""
	}
	f, err := os.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, BufferSize)
	h := md5.New()

	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return GetMd5Sum(h, nil)
}

// GetMd5Sum gets md5 sum as a string and appends the current hash to b.
func GetMd5Sum(md5 hash.Hash, b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}
