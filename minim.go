package main

import (
    "crypto/md5"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    // "io/ioutil"
)

type FileUnit string
type Option int
type OptionRegistration (func(r *RequestOptions, arg string) error)

const (
    B  FileUnit = "B"
    KB FileUnit = "KB"
    MB FileUnit = "MB"
    GB FileUnit = "GB"
)

var RegisteredOptions map[string]OptionRegistration
var FileLocks map[string]*sync.Mutex
var FileLocksLock sync.Mutex

const Name string = "mimime"

var (
    HomePath      string
    CachePath     string
    CacheOrigPath string
    CacheRedPath  string
)

const (
    FileSizeOption    Option = iota
    SslOption         Option = iota
    ForceReloadOption Option = iota
)

var (
    DefaultFileSize FileSize = FileSize{50, KB}
)

type FileSize struct {
    value float64
    unit  FileUnit
}

type RequestOptions struct {
    setOpts map[Option]bool
    fs      FileSize
}

type Request struct {
    imgUrl  string
    imgId   string
    reqOpts RequestOptions
}

func (fs FileSize) String() string {
    return fmt.Sprintf("%f%s", fs.value, fs.unit)
}

func (r *Request) GImgId() string {
    if r.imgId == "" {
        r.imgId = fmt.Sprintf("%x", md5.Sum([]byte(r.imgUrl)))
    }
    return r.imgId
}

func (r *Request) RedPath() string {
    return filepath.Join(
        CacheRedPath,
        r.GImgId()+"-"+r.reqOpts.fs.String()+".jpg")
}

func (r *Request) OrigPath() string {
    return filepath.Join(CacheOrigPath, r.GImgId())
}

func init() {
    HomePath = os.Getenv("HOME")
    CachePath = filepath.Join(HomePath, ".cache", Name)
    CacheOrigPath = filepath.Join(CachePath, "orig")
    CacheRedPath = filepath.Join(CachePath, "red")

    err := os.MkdirAll(HomePath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }
    err = os.MkdirAll(CachePath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }
    err = os.MkdirAll(CacheOrigPath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }
    err = os.MkdirAll(CacheRedPath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }

    RegisteredOptions = make(map[string]OptionRegistration)
    FileLocks = make(map[string]*sync.Mutex)
    addOption([]string{"s", "-size", "-filesize"}, RegisterFileSizeOption)
    addOption([]string{"p", "-ssl"}, GenToggleOptionRegistration(SslOption))
    addOption([]string{"f", "-force"}, GenToggleOptionRegistration(ForceReloadOption))
}

func LogErr(w io.Writer, err error) {
    fmt.Fprintf(w, "Fatal error: %s", err)
}

func addOption(prefixes []string, or OptionRegistration) {
    for _, el := range prefixes {
        RegisteredOptions[el] = or
    }
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return true, err
}

func ParseFileUnit(unparsedFileUnit string) (*FileUnit, error) {
    var fu FileUnit
    switch strings.ToLower(unparsedFileUnit) {
    case "b":
        fu = B
    case "kb":
        fu = KB
    case "mb":
        fu = MB
    case "gb":
        fu = GB
    default:
        fu = KB
    }
    return &fu, nil
}

func ParseFileSize(unparsedFileSize string) (*FileSize, error) {
    splitIndex := len(unparsedFileSize)
    for index, character := range unparsedFileSize {
        if !strings.Contains("1234567890.", string(character)) {
            splitIndex = index
            break
        }
    }
    size, err := strconv.ParseFloat(unparsedFileSize[:splitIndex], 64)
    if err != nil {
        return nil, err
    }
    fu, err := ParseFileUnit(unparsedFileSize[splitIndex:])
    if err != nil {
        return nil, err
    }
    return &FileSize{size, *fu}, nil
}

func RegisterFileSizeOption(r *RequestOptions, arg string) error {
    fs, err := ParseFileSize(arg)
    if err != nil {
        return err
    }
    r.setOpts[FileSizeOption] = true
    r.fs = *fs
    return nil
}

func GenToggleOptionRegistration(opt Option) OptionRegistration {
    return func(r *RequestOptions, arg string) error {
        r.setOpts[opt] = true
        return nil
    }
}

func RegisterOption(ro *RequestOptions, arg string) error {
    for pre, or := range RegisteredOptions {
        if strings.HasPrefix(arg, pre) {
            return or(ro, strings.TrimPrefix(arg, pre))
        }
    }
    errMsg := fmt.Sprintf("Unknown option: %s", arg)
    return errors.New(errMsg)
}

func ParseRequestOptions(unparsedOptions []string) (*RequestOptions, error) {
    setOpts := make(map[Option]bool)
    ro := &RequestOptions{setOpts, DefaultFileSize}
    for _, el := range unparsedOptions {
        err := RegisterOption(ro, el)
        if err != nil {
            return nil, err
        }
    }
    return ro, nil
}

func ParseRequest(path string) (*Request, error) {
    urlSplit := strings.SplitN(path, "/u", 2)
    var imgUrl string
    var options []string
    if len(urlSplit) > 1 {
        imgUrl = urlSplit[1]
        options = strings.Split(urlSplit[0], "/")
    } else {
        imgUrl = urlSplit[0][1:]
        options = []string{""}
    }
    reqOpts, err := ParseRequestOptions(options[1:])
    if err != nil {
        return nil, err
    }
    return &Request{imgUrl, "", *reqOpts}, nil
}

func LogRequest(req *Request) {
    fmt.Printf("Requesting image %s\n", req.imgUrl)

    sslString := "disabled"
    if req.reqOpts.setOpts[SslOption] {
        sslString = "enabled"
    }
    fmt.Printf("SSL option is %s\n", sslString)
    forceString := "disabled"
    if req.reqOpts.setOpts[ForceReloadOption] {
        forceString = "enabled"
    }
    fmt.Printf("Force reload option is %s\n", forceString)

    if req.reqOpts.setOpts[FileSizeOption] {
        fmt.Printf("Option FileSize is set with value: %s\n", req.reqOpts.fs)
    } else {
        fmt.Printf("No FileSize set using default value: %s\n", req.reqOpts.fs)
    }
}

func WriteRequestResponse(req *Request, path string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    protocol := "http"
    if req.reqOpts.setOpts[SslOption] {
        protocol = "https"
    }

    response, err := http.Get(protocol + "://" + req.imgUrl)
    if err != nil {
        return err
    }
    defer response.Body.Close()

    _, err = io.Copy(f, response.Body)
    if err != nil {
        return err
    }

    return nil
}

func CreateOriginalFile(req *Request) error {
    path := filepath.Join(CacheOrigPath, req.GImgId())
    CoLock(req.GImgId())
    defer CoUnlock(req.GImgId())

    if !req.reqOpts.setOpts[ForceReloadOption] {
        ex, err := exists(path)
        if err != nil {
            return err
        }
        if ex {
            return nil
        }
    }

    return WriteRequestResponse(req, path)
}

func CoLock(key string) {
    FileLocksLock.Lock()
    lock, ok := FileLocks[key]
    if !ok {
        lock = &sync.Mutex{}
        FileLocks[key] = lock
    }
    lock.Lock()
    FileLocksLock.Unlock()
}

func CoUnlock(key string) {
    FileLocksLock.Lock()
    FileLocks[key].Unlock()
    FileLocksLock.Unlock()
}

func Minify(req *Request) error {
    cmd := exec.Command(
        "convert",
        "-define",
        "jpeg:extent="+req.reqOpts.fs.String(),
        req.OrigPath(),
        req.RedPath())
    return cmd.Run()
}

func handler(w http.ResponseWriter, r *http.Request) {
    req, err := ParseRequest(r.URL.Path)
    if err != nil {
        LogErr(w, err)
        return
    }
    go LogRequest(req)

    err = CreateOriginalFile(req)
    if err != nil {
        LogErr(w, err)
        return
    }

    err = Minify(req)
    if err != nil {
        LogErr(w, err)
        return
    }

    f, err := os.Open(req.RedPath())
    defer f.Close()
    if err != nil {
        LogErr(w, err)
        return
    }

    _, err = io.Copy(w, f)
    if err != nil {
        LogErr(w, err)
        return
    }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
