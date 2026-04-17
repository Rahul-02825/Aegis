package loadBalancer


// factory design pattern implementation for the balancers
type Loadbalancer interface{
	insertServer(server string)
	removeServer(server string)	
	getServer(request string) string
}
// for now consistent hashing is only available balancer
func balancerFactory(balancerType string) (Loadbalancer,error){

	switch balancerType{
	case "consistent-hash":
			return &ConsistentHash{},nil
		
	default:
		return nil,nil
	}	
	
}

