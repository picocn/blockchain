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

	"strconv"
	"strings"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
const BANK = 6
const MANUFACTURER = 10
const DEALER = 9
const LOGISTICS = 11

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
	Factory        string  `json:"factory"`
	Model          string  `json:"model"`
	Colour         string  `json:"colour"`
	CarID          string  `json:"CarID"`
	Price          float64 `json:"price"`
	Dealer         string  `json:"dealer"`
	Holder         string  `json:"holder"`
	Status         int     `json:"status"`
	OrderID        string  `json:"OrderID"`
	Location       string  `json:"location"`
	LoanBank       string  `json:"bank"`
	LoanContractID string  `json:"loanContractID"`
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

func (t *SimpleChaincode) create_Order(stub *shim.ChaincodeStub, caller string, caller_affiliation int, args []string) ([]byte, error) {

	//var v Vehicle
	orderid := caller + "_" + args[5]
	Factory := "\"Factory\":\"" + args[0] + "\", " // Variables to define the JSON
	Model := "\"Model\":\"" + args[1] + "\","
	Colour := "\"Colour\":\"" + args[2] + "\", "
	CarID := "\"CarID\":\"UNDEFINED\", "

	Price := "\"Price\":\"" + args[3] + "\", "
	Dealer := "\"Dealer\":\"" + caller + "\", "
	Holder := "\"Holder\":\"" + args[4] + "\", "
	Status := "\"Status\":\"INIT\", "
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
	err = json.Unmarshal([]byte(vehicle_json), &v) // Convert the JSON defined above into a vehicle object for go

	if err != nil {
		return nil, errors.New("Invalid JSON object")
	}

	if caller_affiliation != DEALER { // Only the dealer can create a new order

		return nil, errors.New("Only dealers can make new order!")
	}

	_, err = t.save_changes(stub, v)

	if err != nil {
		fmt.Printf("CREATE_ORDER: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	bytes, err := stub.GetState("OrderIDs")

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)

	if err != nil {
		return nil, errors.New("Corrupt Order_Holder record")
	}

	OrderIDs.Orders = append(OrderIDs.Orders, OrderID)

	bytes, err = json.Marshal(OrderIDs)

	if err != nil {
		fmt.Print("Error creating Order_Holder record")
	}

	err = stub.PutState("OrderIDs", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================
func (t *SimpleChaincode) get_vehicle_details(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int) ([]byte, error) {

	bytes, err := json.Marshal(v)

	if err != nil {
		return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object")
	}

	if v.Holder == caller ||
		caller_affiliation == BANK {

		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied")
	}

}

//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================

func (t *SimpleChaincode) get_vehicles(stub *shim.ChaincodeStub, caller string, caller_affiliation int) ([]byte, error) {

	bytes, err := stub.GetState("OrderIDs")

	if err != nil {
		return nil, errors.New("Unable to get OrderIDs")
	}

	var OrderIDs Order_Holder

	err = json.Unmarshal(bytes, &OrderIDs)

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

	caller, caller_affiliation, err := t.get_caller_data(stub)

	if err != nil {
		fmt.Printf("QUERY: Error retrieving caller details", err)
		return nil, errors.New("QUERY: Error retrieving caller details")
	}

	if function == "get_vehicle_details" {

		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed")
			return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		v, err := t.retrieve_car(stub, args[0])
		if err != nil {
			fmt.Printf("QUERY: Error retrieving v5c: %s", err)
			return nil, errors.New("QUERY: Error retrieving car " + err.Error())
		}

		return t.get_vehicle_details(stub, v, caller, caller_affiliation)

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
//        args<transfer>: 0:order_id, transfer_receipient
//        args<update>:   0:order_id, new_value
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub)

	if err != nil {
		return nil, errors.New("Error retrieving caller information")
	}

	if function == "create_Order" {
		return t.create_Order(stub, caller, caller_affiliation, args)
	} else if function == "update_loc" {
		return t.update_loc(stub, caller, caller_affiliation, args[0], args[1])
	} else if function == "update_state_repayment" {
		return t.update_state_repayment(stub, caller, caller_affiliation, args[0])
	} else { // If the function is not a create then there must be a car so we need to retrieve the car.

		v, err := t.retrieve_car(stub, args[0])

		if err != nil {
			fmt.Printf("INVOKE: Error retrieving order: %s", err)
			return nil, errors.New("Error retrieving order")
		}

		ecert, err := t.get_ecert(stub, args[1])

		if err != nil {
			return nil, err
		}
		rec_affiliation, err := t.check_affiliation(stub, string(ecert))

		if err != nil {
			return nil, err
		}
		if function == "bank_confirm_order" {
			return t.bank_confirm_order(stub, v, caller, caller_affiliation, args[0], rec_affiliation)
		} else if function == "bank_confirm_deliver" {
			return t.bank_confirm_deliver(stub, v, caller, caller_affiliation, args[0], rec_affiliation)
		} else if function == "manufacturer_deliver" {
			return t.manufacturer_deliver(stub, v, caller, caller_affiliation, args[0], rec_affiliation)
		} else if function == "logistics_deliver" {
			return t.logistics_deliver(stub, v, caller, caller_affiliation, args[0], rec_affiliation)
		}

		return nil, errors.New("Function of that name doesn't exist.")

	}
}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 bank confirm
//=================================================================================================================================
func (t *SimpleChaincode) bank_confirm_order(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {

	if v.Status == STATE_INIT &&
		v.Holder == caller &&
		caller_affiliation == BANK &&
		recipient_affiliation == MANUFACTURER { // If the roles and users are ok

		v.Holder = recipient_name    // then make the owner the new owner
		v.Status = STATE_MANUFACTURE // and mark it in the state of manufacture

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
func (t *SimpleChaincode) manufacturer_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {

	if v.Status == STATE_MANUFACTURE &&
		v.Holder == caller &&
		caller_affiliation == MANUFACTURER &&
		recipient_affiliation == LOGISTICS { // If the roles and users are ok
		v.Holder = recipient_name // then make the owner the new owner
		v.Status = STATE_SHIPPING // and mark it in the state of DELIVER

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
func (t *SimpleChaincode) update_loc(stub *shim.ChaincodeStub, caller string, caller_affiliation int, orderID string, location string) ([]byte, error) {
	v, err := t.retrieve_car(stub, orderID)
	if err != nil {
		fmt.Printf("INVOKE: Error retrieving order: %s", err)
		return nil, errors.New("Error retrieving order")
	}
	if v.Status == STATE_SHIPPING &&
		v.Holder == caller &&
		caller_affiliation == LOGISTICS { // If the roles and users are ok
		v.Location = location // then make the owner the new owner
		_, err = t.save_changes(stub, v)
		if err != nil {
			fmt.Printf("Cannot update location")
			return nil, errors.New("Cannot update location")
		}

	} else { // Otherwise if there is an error

		fmt.Printf("Update location: Permission Denied")
		return nil, errors.New("Permission Denied")

	}

	return nil, nil // We are Done
}

func (t *SimpleChaincode) update_state_repayment(stub *shim.ChaincodeStub, caller string, caller_affiliation int, orderID string) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) bank_confirm_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	return nil, nil
}

func (t *SimpleChaincode) logistics_deliver(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	return nil, nil
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

	err = json.Unmarshal(bytes, &v)

	if err != nil {
		fmt.Printf("RETRIEVE_CAR: Corrupt vehicle record "+string(bytes)+": %s", err)
		return v, errors.New("RETRIEVE_CAR: Corrupt vehicle record" + string(bytes))
	}

	return v, nil
}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub *shim.ChaincodeStub) (string, int, error) {

	user, err := t.get_username(stub)
	if err != nil {
		return "", -1, err
	}

	ecert, err := t.get_ecert(stub, user)
	if err != nil {
		return "", -1, err
	}

	affiliation, err := t.check_affiliation(stub, string(ecert))
	if err != nil {
		return "", -1, err
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

	fmt.Println("HTTP RESPONSE", response)

	if err != nil {
		return nil, errors.New("Error calling ecert API")
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body) // Read the response from the http callout into the variable contents

	fmt.Println("HTTP BODY:", string(contents))

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

	return x509Cert.Subject.CommonName, nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub *shim.ChaincodeStub, cert string) (int, error) {

	decodedCert, err := url.QueryUnescape(cert) // make % etc normal //

	if err != nil {
		return -1, errors.New("Could not decode certificate")
	}

	pem, _ := pem.Decode([]byte(decodedCert)) // Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes) // Extract Certificate from argument //

	if err != nil {
		return -1, errors.New("Couldn't parse certificate")
	}

	cn := x509Cert.Subject.CommonName

	res := strings.Split(cn, "\\")

	affiliation, _ := strconv.Atoi(res[2])

	return affiliation, nil
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
