namespace go user

struct UserInfo {
    1: i64 user_id
    2: string username
    3: string avatar_url
    4: string created_at
}

struct RegisterRequest {
    1: string username
    2: string password
}

struct LoginRequest {
    1: string username
    2: string password
}

service UserService {
    UserInfo GetUserInfo(1: i64 user_id)
    i64 RegisterUser(1: RegisterRequest req)
    string Login(1: LoginRequest req)
    bool CheckToken(1: string token)
}