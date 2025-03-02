namespace go order

struct UpdateReq {
    1: string orderID
    2: string status
}

service OrderService {
    bool HealthCheck()
    bool UpdateStatus(1: UpdateReq req)
}