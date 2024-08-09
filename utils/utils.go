package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Utilities interface {
	ConvertToInt64(str string) (int64, error)
	ReadIntParam(c echo.Context, str string) (int, error)
	ReadStringParam(c echo.Context, str string) (string, error)
	ReadJSON(c echo.Context, dst interface{}) error
	ReadFormData(c echo.Context, dst interface{}) error
	ReadStringQuery(qs url.Values, key string, defaultValue string) string
	ReadIntQuery(qs url.Values, key string, defaultValue int) int
	HandleFiles(c echo.Context, key, name string) ([]string, error)
	GenerateSignature(orderId, id, secret string) string
	MakeCustomRequest(httpClient *http.Client, req *http.Request) (map[string]interface{}, error)
	AddHeaderIfMissing(w http.ResponseWriter, key, value string)

	InternalServerError(c echo.Context, err error)
	BadRequest(c echo.Context, err error)
	MethodNotFound(c echo.Context)
	NotFoundResponse(c echo.Context)
	EditConflictResponse(c echo.Context)
	UserUnAuthorizedResponse(c echo.Context, err error)
	RateLimitExceededResponse(c echo.Context)
	CustomErrorResponse(c echo.Context, message Cake, status int, err error)
	ValidationError(c echo.Context, err error)
}

type utilsImpl struct {
}

var uploadDir string = "./public"

func NewUtils() Utilities {
	return &utilsImpl{}
}

func (u *utilsImpl) ConvertToInt64(str string) (int64, error) {
	integer, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return integer, nil
}

func (u *utilsImpl) ReadIntParam(c echo.Context, str string) (int, error) {
	param := c.Param(str)
	id, err := strconv.Atoi(param)
	if err != nil || id < 1 {
		return 0, errors.New("invalid parameter")
	}

	return id, err
}

func (u *utilsImpl) ReadStringParam(c echo.Context, str string) (string, error) {
	param := c.Param(str)
	if param == "" {
		return "", errors.New("invalid parameter")
	}
	return param, nil
}

func (u *utilsImpl) ReadJSON(c echo.Context, dst interface{}) error {
	maxBytes := 1_048_576
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, int64(maxBytes))

	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (u *utilsImpl) ReadFormData(c echo.Context, dst interface{}) error {
	err := c.Request().ParseMultipartForm(10 << 20)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %v", err)
	}
	fmt.Println(c.FormValue("images"), "\nc.FormValue(images)")

	dstValue := reflect.ValueOf(dst).Elem()
	dstType := dstValue.Type()

	for i := 0; i < dstValue.NumField(); i++ {
		field := dstType.Field(i)
		fieldValue := dstValue.Field(i)
		formValue := c.FormValue(strings.ToLower(field.Name))

		if fieldValue.CanSet() {
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(formValue)
			case reflect.Int64:
				if formValue == "" {
					fieldValue.SetInt(0)
				} else {
					value, err := u.ConvertToInt64(formValue)
					if err != nil {
						return fmt.Errorf("invalid value for field %s: %v", field.Name, err)
					}
					fieldValue.SetInt(value)
				}
			}
		}
	}

	return nil
}

func (u *utilsImpl) ReadStringQuery(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func (u *utilsImpl) ReadIntQuery(qs url.Values, key string, defaultValue int) int {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	res, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return res
}

func (u *utilsImpl) HandleFiles(c echo.Context, key string, name string) ([]string, error) {
	files := c.Request().MultipartForm.File[key]
	fmt.Println(files, "\n\nfiles")
	if len(files) == 0 {
		return []string{}, nil
	}
	fmt.Println("\n\nHandleFiles")
	filePaths := []string{}
	uploadDir := uploadDir + "/" + key
	fmt.Println(uploadDir, "uploadDir")

	for _, fileHeader := range files {
		file, err := fileHeader.Open()

		if err != nil {
			return []string{}, err
		}
		defer file.Close()

		b := make([]byte, 4)
		rand.Read(b)
		suffix := hex.EncodeToString(b)
		filename := fmt.Sprintf("%s_%s%s", name, suffix, filepath.Ext(fileHeader.Filename))

		dst, err := os.Create(filepath.Join(uploadDir, filename))
		if err != nil {
			return []string{}, err
		}
		defer dst.Close()
		filePaths = append(filePaths, uploadDir[1:]+"/"+filename)

		_, err = io.Copy(dst, file)
		if err != nil {
			return []string{}, err
		}
	}
	return filePaths, nil
}

func (u *utilsImpl) GenerateSignature(orderId, id, secret string) string {
	data := orderId + "|" + id
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (u *utilsImpl) MakeCustomRequest(httpClient *http.Client, req *http.Request) (map[string]interface{}, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var data Cake
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *utilsImpl) AddHeaderIfMissing(w http.ResponseWriter, key, value string) {
	for _, h := range w.Header()[key] {
		if h == value {
			return
		}
	}
	w.Header().Add(key, value)
}
