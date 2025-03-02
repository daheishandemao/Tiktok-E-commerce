namespace go order

service OrderService {
    string UpdateStatus(1: string orderID, 2: string status)
}