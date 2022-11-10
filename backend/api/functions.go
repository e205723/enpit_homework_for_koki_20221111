package api

import (
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "time"
    "github.com/dgrijalva/jwt-go"
    _ "github.com/lib/pq"
)

type Server struct {
    Db *sql.DB
}

type Response struct {
    Message string `json:"message"`
}

type SignUpRequest struct {
    Name                   string `json:"name"`
    Password               string `json:"password"`
    PasswordConfirmination string `json:"passwordConfirmination"`
}

type SignUpResponse struct {
    Name string `json:"name"`
}

type Claims struct {
    Name string
    jwt.StandardClaims
}

var charset62 = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var jwtKey = []byte(RandomString(511))

func RandomString(length int) string {
    randomString := make([]rune, length)
    for i := range randomString {
        randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset62))))
        if err != nil {
            log.Println(err)
            return ""
        }
        randomString[i] = charset62[int(randomNumber.Int64())]
    }
    return string(randomString)
}

func SetJwtInCookie(w http.ResponseWriter, userName string) {
    expirationTime := time.Now().Add(672 * time.Hour)
    claims := &Claims{
        Name: userName,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    cookie := &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    }
    http.SetCookie(w, cookie)
}

func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
        // 問題1↓は何してる? (使う用語: 構造体、初期化、パース)
    var signUpRequest SignUpRequest
    decoder := json.NewDecoder(r.Body)
    decodeError := decoder.Decode(&signUpRequest)
        // 問題1終わり
    if decodeError != nil {
        log.Println("[ERROR]", decodeError)
    }
    if signUpRequest.Password != signUpRequest.PasswordConfirmination {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
        // 問題2↓パスワードをハッシュ関数でハッシュ化させてDBに保存している、なぜ?
        //   ハッシュ関数の特性を3つ挙げて答えよ
    passwordHash32Byte := sha256.Sum256([]byte(signUpRequest.Password))
    passwordHashURLSafe := base64.URLEncoding.EncodeToString(passwordHash32Byte[:])
    queryToReGisterUser := fmt.Sprintf("INSERT INTO users (name, password_hash) VALUES ('%s', '%s')", signUpRequest.Name, passwordHashURLSafe)
    _, queryRrror := s.Db.Exec(queryToReGisterUser)
        // 問題2終わり
        if queryRrror != nil {
                log.Println("[ERROR]", queryRrror)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
        // 問題3
        // SetJwtInCookie関数ではJWTトークンを生成してCookieに設定している
        // 参考にしたサイト(英): https://www.sohamkamani.com/golang/jwt-authentication/
        // 日本語の良いサイト: https://moneyforward.com/engineers_blog/2020/09/15/jwt/
        // JWTの仕組みを勉強してみよう
        // ちなみにDBにパスワードを保存するときに使うハッシュ関数はJWTトークンを生成するときも使われる
        // 問題3-1 JWTのメリットはなに?(使う用語: SPOF, スケール)
        // 問題3-2 JWTのセキュリティ上のデメリットとして、トークンに使う秘密鍵がハッカーにバレるとどのようなリスクがある?(使う用語: 全員、ログイン)
    SetJwtInCookie(w, signUpRequest.Name)
        // 問題3終わり(今度またJWTは勉強しよう、内容が今回ちょっと重いので実装だけにする)
    w.Header().Set("Content-Type", "application/json")
    response := SignUpResponse{
        Name: signUpRequest.Name,
    }
    jsonResponse, err := json.Marshal(response)
    if err != nil {
        log.Fatalf("Error happened in JSON marshal. Err: %s", err)
    }
    w.Write(jsonResponse)
}
