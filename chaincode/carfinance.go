package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
const BANK = "cib"
const MANUFACTURER = "manufacturer"
const DEALER = "dealer"
const LOGISTICS = "logistics"

//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can
//					be done to the vehicle at points in it's lifecycle
//==============================================================================================================================
const STATE_INIT = 0
const STATE_LOANTRANSFERED = 1
const STATE_MANUFACTURE = 2
const STATE_SHIPPING = 3
const STATE_LOANRETURNED = 4
const STATE_FINISHED = 5

type SimpleChaincode struct {
}

//==============================================================================================================================
//	ECertResponse - Struct for storing the JSON response of retrieving an ECert. JSON OK -> Struct OK
//==============================================================================================================================
type ECertResponse struct {
	OK    string `json:"OK"`
	Error string `json:"Error"`
}

//==============================================================================================================================
//	Vehicle - Defines the structure for a car object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Vehicle struct {
	Factory        string  `json:"Factory"`
	Model          string  `json:"Model"`
	Colour         string  `json:"Colour"`
	CarID          string  `json:"CarID"`
	Price          float64 `json:"Price"`
	Dealer         string  `json:"Dealer"`
	Holder         string  `json:"Holder"`
	Status         int     `json:"Status"`
	OrderID        string  `json:"OrderID"`
	Location       string  `json:"Location"`
	LoanBank       string  `json:"Loanbank"`
	LoanContractID string  `json:"LoanContractID"`
}

type Order_Car struct {
	OrderID string `json:"OrderID"`
	CarID   string `json:"CarID"`
}

//==============================================================================================================================
//	Cars Holder - Defines the structure that holds all the CarIDs for vehicles that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type Car_Holder struct {
	Cars []Order_Car `json:"Cars"`
}

type Order_Holder struct {
	Orders []string `json:"Orders"`
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub *shim.ChaincodeStub, v Vehicle) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error converting vehicle record: %s", err)
		return false, errors.New("Error converting vehicle record")
	}

	err = stub.PutState(v.OrderID, bytes)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error storing vehicle record: %s", err)
		return false, errors.New("Error storing vehicle record")
	}

	return true, nil
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Vehicle - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================

func (t *SimpleChaincode) create_Order(stub *shim.ChaincodeStub, caller string, caller_affiliation string, args []string) ([]byte, error) {

	//var v Vehicle
	orderid := caller + "_" + args[5]
	Factory := "\"Factory\":\"" + args[1] + "\", " // Variables to define the JSON
	Model := "\"Model\":\"" + args[2] + "\","
	Colour := "\"Colour\":\"" + args[3] + "\", "
	CarID := "\"CarID\":\"UNDEFINED\", "

	Price := "\"Price\":" + args[4] + ", "
	Dealer := "\"Dealer\":\"" + caller + "\", "
	Holder := "\"Holder\":\"" + args[6] + "\", "
	Status := "\"Status\": 0, "
	OrderID := "\"OrderID\":\"" + orderid + "\", "
	LoanBank := "\"LoanBank\":\"" + args[6] + "\","
	LoanContractID := "\"LoanContractID\":\"" + args[7] + "\""

	vehicle_json := "{" + Factory + Model + Colour + CarID + Price + Dealer + Holder + Status + OrderID + LoanBank + LoanContractID + "}" // Concatenates the variables to create the total JSON object

	if LoanContractID == "" {
		return nil, errors.New("Invalid LoanContractID provided")
	}

	car, err := t.retrieve_car(stub, orderid)
	if err != nil {
		return nil, errors.New("error get orderid!")
	} else {
		if car.OrderID == orderid {
			return nil, errors.New("OrderID exists!")
		}
	}

	var v Vehicle
	fmt.Println(vehicle_json)
	err = json.Unmarshal([]byte(vehicle_json), &v) // Convert the JSON defined above into a vehicle object for go

	if err != nil {
		return nil, errors.New("Convert json to struct: Invalid JSON object")
	}
	fmt.Printf("affiliation: %d\n", caller_affiliation)
	if caller_affiliation != DEALER { // Only the dealer can create a new order

		return nil, errors.New("Only dealers can make new order!")
	}

	_, err = t.save_changes(stub, v)

	if err != nil {
		fmt.Printf("CREATE_ORDER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	bytes, err := stub.GetState("OrderIDs")
	fmt.Println(bytes)

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)

	if err != nil {
		return nil, errors.New("Corrupt Order_Holder record")
	}

	OrderIDs.Orders = append(OrderIDs.Orders, orderid)

	bytes, err = json.Marshal(OrderIDs)

	if err != nil {
		fmt.Print("Error creating Order_Holder record")
	}
	fmt.Println("putting orderids ", string(bytes))
	err = stub.PutState("OrderIDs", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

//=================================================================================================================================
//	 get_vehicle_details only holder or dealer to find requested order
//=================================================================================================================================
func (t *SimpleChaincode) get_vehicle_details(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := json.Marshal(v)

	if err != nil {
		return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object")
	}

	if v.Holder == caller ||
		(v.Dealer == caller &&
			v.Status == STATE_INIT) {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied")
	}

}

//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================

func (t *SimpleChaincode) get_vehicles(stub *shim.ChaincodeStub, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := stub.GetState("OrderIDs")

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)
	fmt.Println(OrderIDs)

	if err != nil {
		return nil, errors.New("Corrupt Order_Holder")
	}

	result := "["

	var temp []byte
	var v Vehicle

	for _, OrderID := range OrderIDs.Orders {

		v, err = t.retrieve_car(stub, OrderID)

		if err != nil {
			return nil, errors.New("Failed to retrieve car")
		}

		temp, err = t.get_vehicle_details(stub, v, caller, caller_affiliation)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	//Args
	//				0
	//			peer_address

	var CarIDs Car_Holder
	var OrderIDs Order_Holder

	bytes, err := json.Marshal(CarIDs)

	if err != nil {
		return nil, errors.New("Error creating Car_Holder record")
	}

	err = stub.PutState("CarIDs", bytes)

	orders, err := json.Marshal(OrderIDs)

	if err != nil {
		return nil, errors.New("Error creating Order_Holder record")
	}
	err = stub.PutState("OrderIDs", orders)

	err = stub.PutState("Peer_Address", []byte(args[0]))
	if err != nil {
		return nil, errors.New("Error storing peer address")
	}

	return nil, nil
}

//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub, args[0])

	if err != nil {
		fmt.Printf("QUERY: Error retrieving caller details", err)
		return nil, errors.New("QUERY: Error retrieving caller details")
	}

	if function == "get_vehicle_details" {

		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments passed")
			return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		v, err := t.retrieve_car(stub, args[1])
		if err != nil {
			fmt.Printf("QUERY: Error retrieving v5c: %s", err)
			return nil, errors.New("QUERY: Error retrieving car " + err.Error())
		}

		return t.get_vehicle_details(stub, v, caller, caller_affiliation)

	} else if function == "get_all_vehicles" {
		return t.get_all_vehicles(stub, caller, caller_affiliation)
	} else if function == "get_repay_vehicles" {

		return t.get_repay_vehicles(stub, caller, caller_affiliation)

	} else if function == "get_vehicles" {
		return t.get_vehicles(stub, caller, caller_affiliation)
	}
	return nil, errors.New("Received unknown function invocation")
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//
//		  create_order: args: orderid,...
//
//        args<transfer>: 1:order_id, transfer_receipient
//        args<update>:   1:order_id, new_value
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub, args[0])

	if err != nil {
		return nil, errors.New("Error retrieving caller information")
	}

	if function == "create_Order" {
		return t.create_Order(stub, caller, caller_affiliation, args)
	} else { // If the function is not a create then there must be a car so we need to retrieve the car.

		v, err := t.retrieve_car(stub, args[1])
		fmt.Printf("retrieved car: %s, holder=%s\n", v.OrderID, v.Holder)
		if err != nil {
			fmt.Printf("INVOKE: Error retrieving order: %s", err)
			return nil, errors.New("Error retrieving order")
		}
		var rec_affiliation string
		if strings.Contains(function, "update") == false {
			ecert, err := t.get_ecert(stub, args[2])

			if err != nil {
				return nil, err
			}
			rec_affiliation, err = t.check_affiliation(stub, string(ecert))

			if err != nil {
				return nil, err
			}
		}
		if function == "bank_confirm_order" {
			return t.bank_confirm_order(stub, v, caller, caller_affiliation, args[2], rec_affiliation)
		} else if function == "bank_confirm_deliver" {
			return t.bank_confirm_deliver(stub, v, caller, caller_affiliation, args[2], rec_affiliation)
		} else if function == "manufacturer_deliver" {
			return t.manufacturer_deliver(stub, v, caller, caller_affiliation, args[2], rec_affiliation, args[3])
		} else if function == "logistics_deliver" {
			return t.logistics_deliver(stub, v, caller, caller_affiliation, v.Dealer, rec_affiliation)
		} else if function == "update_state_repayment" {
			return t.update_state_repayment(stub, v, caller, caller_affiliation)
		} else if function == "update_loc" {
			return t.update_loc(stub, v, caller, caller_affiliation, args[2])
		}

		return nil, errors.New("Function of that name doesn't exist.")

	}
}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 bank confirm
//=================================================================================================================================
func (t *SimpleChaincode) bank_confirm_order(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {
	fmt.Printf("bank_confirm: %d,holder=%s,caller_aff=%s,recip_aff=%s\n", v.Status, v.LoanBank, caller_affiliation, recipient_affiliation)
	fmt.Println(v.Colour)
	if v.Status == STATE_INIT &&
		v.LoanBank == caller &&
		caller_affiliation == BANK &&
		recipient_affiliation == MANUFACTURER { // If the roles and users are ok

		v.Holder = recipient_name       // then make the owner the new owner
		v.Status = STATE_LOANTRANSFERED // and mark it in the state of manufacture

	} else { // Otherwise if there is an error

		fmt.Printf("BANK_CONFIRM: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	_, err := t.save_changes(stub, v) // Write new state

	if err != nil {
		fmt.Printf("BAKN_CONFIRM: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil // We are Done

}

//=================================================================================================================================
//	 manufacturer deliver
//=================================================================================================================================
func (t *SimpleChaincode) manufacturer_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string, carID string) ([]byte, error) {
	fmt.Printf("MANUFACTURER DELIVER: carid=%s\n", carID)
	if v.Status == STATE_LOANTRANSFERED &&
		v.Holder == caller &&
		caller_affiliation == MANUFACTURER &&
		recipient_affiliation == LOGISTICS { // If the roles and users are ok
		v.Holder = recipient_name // then make the owner the new owner
		v.Status = STATE_SHIPPING // and mark it in the state of DELIVER
		v.CarID = carID           //add CAR ID
	} else { // Otherwise if there is an error

		fmt.Printf("DELIVER: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	_, err := t.save_changes(stub, v) // Write new state

	if err != nil {
		fmt.Printf("DELIVER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil // We are Done
}

//=====update location=================
//
//=================================================================================================================================
//
//=================================================================================================================================
func (t *SimpleChaincode) update_loc(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string, location string) ([]byte, error) {
	if v.Status == STATE_SHIPPING &&
		v.Holder == caller &&
		caller_affiliation == LOGISTICS { // If the roles and users are ok

		v.Location = location // and mark it in the state of DELIVER

	} else { // Otherwise if there is an error

		fmt.Printf("DELIVER: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	_, err := t.save_changes(stub, v) // Write new state

	if err != nil {
		fmt.Printf("DELIVER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil // We are Done
}

func (t *SimpleChaincode) update_state_repayment(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {
	if v.Status == STATE_SHIPPING &&
		v.LoanBank == caller &&
		caller_affiliation == BANK { // If the roles and users are ok

		v.Status = STATE_LOANRETURNED // and mark it in the state of DELIVER

	} else { // Otherwise if there is an error

		fmt.Printf("DELIVER: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	_, err := t.save_changes(stub, v) // Write new state

	if err != nil {
		fmt.Printf("DELIVER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil // We are Done
}

func (t *SimpleChaincode) bank_confirm_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) logistics_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {
	if v.Status == STATE_LOANRETURNED &&
		v.Holder == caller &&
		caller_affiliation == LOGISTICS &&
		recipient_name == v.Dealer && //destination should be the dealer made order
		recipient_affiliation == DEALER { // If the roles and users are ok
		v.Holder = recipient_name // then make the owner the new owner
		v.Status = STATE_FINISHED // and mark it in the state of DELIVER

	} else { // Otherwise if there is an error

		fmt.Printf("DELIVER: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	_, err := t.save_changes(stub, v) // Write new state

	if err != nil {
		fmt.Printf("DELIVER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	return nil, nil // We are Done
}

//==============================================================================================================================
//	 retrieve_car - Gets the state of the data at carID in the ledger then converts it from the stored
//					JSON into the Vehicle struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_car(stub *shim.ChaincodeStub, OrderID string) (Vehicle, error) {

	var v Vehicle

	bytes, err := stub.GetState(OrderID)

	if err != nil {
		fmt.Printf("RETRIEVE_CAR: Failed to invoke vehicle_code: %s", err)
		return v, errors.New("RETRIEVE_CAR: Error retrieving vehicle with CarID = " + OrderID)
	}
	length := len(bytes)
	fmt.Print("获取订单:")
	fmt.Println(string(bytes))
	if length > 0 {
		err = json.Unmarshal(bytes, &v)

		if err != nil {
			fmt.Printf("RETRIEVE_CAR: Corrupt vehicle record "+string(bytes)+": %s", err)
			return v, errors.New("RETRIEVE_CAR: Corrupt vehicle record" + string(bytes))
		}

		return v, nil
	} else {
		return v, nil
	}
}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub *shim.ChaincodeStub, caller string) (string, string, error) {

	//	user, err := t.get_username(stub)
	//	if err != nil {
	//		return "", -1, err
	//	}
	user := caller
	ecert, err := t.get_ecert(stub, user)
	if err != nil {
		return "", "", err
	}

	affiliation, err := t.check_affiliation(stub, string(ecert))
	if err != nil {
		return "", "", err
	}

	return user, affiliation, nil
}

func (t *SimpleChaincode) get_ecert(stub *shim.ChaincodeStub, name string) ([]byte, error) {

	var cert ECertResponse

	peer_address, err := stub.GetState("Peer_Address")
	if err != nil {
		return nil, errors.New("Error retrieving peer address")
	}

	response, err := http.Get("http://" + string(peer_address) + "/registrar/" + name + "/ecert") // Calls out to the HyperLedger REST API to get the ecert of the user with that name

	//fmt.Println("HTTP RESPONSE", response)

	if err != nil {
		return nil, errors.New("Error calling ecert API")
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body) // Read the response from the http callout into the variable contents

	//fmt.Println("HTTP BODY:", string(contents))

	if err != nil {
		return nil, errors.New("Could not read body")
	}

	err = json.Unmarshal(contents, &cert)

	if err != nil {
		return nil, errors.New("Could not retrieve ecert for user: " + name)
	}

	fmt.Println("CERT OBJECT:", cert)

	if cert.Error != "" {
		fmt.Println("GET ECERT ERRORED: ", cert.Error)
		return nil, errors.New(cert.Error)
	}

	return []byte(string(cert.OK)), nil
}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub *shim.ChaincodeStub) (string, error) {

	bytes, err := stub.GetCallerCertificate()

	if err != nil {
		return "", errors.New("Couldn't retrieve caller certificate")
	}
	x509Cert, err := x509.ParseCertificate(bytes) // Extract Certificate from result of GetCallerCertificate
	if err != nil {
		return "", errors.New("Couldn't parse certificate")
	}
	fmt.Println(x509Cert.Subject)
	return x509Cert.Subject.CommonName, nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub *shim.ChaincodeStub, cert string) (string, error) {

	decodedCert, err := url.QueryUnescape(cert) // make % etc normal //

	if err != nil {
		return "", errors.New("Could not decode certificate")
	}

	pem, _ := pem.Decode([]byte(decodedCert)) // Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes) // Extract Certificate from argument //

	if err != nil {
		return "", errors.New("Couldn't parse certificate")
	}

	cn := x509Cert.Subject.CommonName

	res := strings.Split(cn, "\\")
	fmt.Println("affiliation: ", res)
	affiliation := res[1]
	fmt.Println(affiliation)
	return affiliation, nil
}

//get all vehicles
//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================
func (t *SimpleChaincode) get_all_vehicle_details(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := json.Marshal(v)

	if err != nil {
		return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object")
	}

	if ((v.Status == STATE_LOANTRANSFERED || v.Status == STATE_SHIPPING ||
		v.Status == STATE_LOANRETURNED || v.Status == STATE_FINISHED) &&
		v.Factory == caller) ||
		v.LoanBank == caller ||
		v.Dealer == caller {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied")
	}

}

//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================

func (t *SimpleChaincode) get_all_vehicles(stub *shim.ChaincodeStub, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := stub.GetState("OrderIDs")

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)
	fmt.Println(OrderIDs)

	if err != nil {
		return nil, errors.New("Corrupt Order_Holder")
	}

	result := "["

	var temp []byte
	var v Vehicle

	for _, OrderID := range OrderIDs.Orders {

		v, err = t.retrieve_car(stub, OrderID)

		if err != nil {
			return nil, errors.New("Failed to retrieve car")
		}

		temp, err = t.get_all_vehicle_details(stub, v, caller, caller_affiliation)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//repayment list
func (t *SimpleChaincode) get_repay_vehicle_details(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := json.Marshal(v)

	if err != nil {
		return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object")
	}

	if (v.LoanBank == caller) &&
		(v.Status == STATE_SHIPPING) {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied")
	}

}

//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================

func (t *SimpleChaincode) get_repay_vehicles(stub *shim.ChaincodeStub, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := stub.GetState("OrderIDs")

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)
	fmt.Println(OrderIDs)

	if err != nil {
		return nil, errors.New("Corrupt Order_Holder")
	}

	result := "["

	var temp []byte
	var v Vehicle

	for _, OrderID := range OrderIDs.Orders {

		v, err = t.retrieve_car(stub, OrderID)

		if err != nil {
			return nil, errors.New("Failed to retrieve car")
		}

		temp, err = t.get_repay_vehicle_details(stub, v, caller, caller_affiliation)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))

	if err != nil {
		fmt.Printf("Error starting Chaincode: %s", err)
	}
}
