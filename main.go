package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Timestamp struct {
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

type Input struct {
	Steering float32 // Steerin response from user, -1.0 - 1.0 in range.
	Throttle float32 // Throttle response from user, 0.0 - 1.0 in range.
	Brake    float32 // Brake response from user, 0.0 - 1.0 in range.
	Clutch   float32 // Clutch response from user, 0.0 - 1.0 in range.
}

// Every json object comming from the simulator contains these two fields.
type BaseData struct {
	Type      string    // Type of data from simulator, can be "Event", "Stream", "ExerciseStart" or "ExerciseEnd"
	Timestamp Timestamp // When the data was generated.
}

// If Type == "Stream" this is the data sent from the simulator
type StreamData struct {
	BaseData
	Speed           int     // Vehicle speed in km/h
	SpeedLimit      int     // Current speed limit in km/h
	FuelConsumption float32 // Current fuel consumption of the vehicle in l/100km
	TurnIndicator   int     // Current turn signal, -1: Turn signal left, 1: Turn signal right, 0: Turn signal off, 2: Warning signals
	Input           Input   // User input to the vehicle
}

// If Type == "Event" this is the data send from the simulator
//
// Event field can contains on of the following values:
// 	BlueLightsEnabled 				- Turn on blue lights on vehicle
// 	BlueLightsDisabled 				- Turned off blue lights on vehicle
//  ForgotTurnIndicator
//  MolestedWildlife				- Drove too fast near wildlife
//  EncounterWildlife				- Driver encounters wildlife, at this point the driver should be able to see and react to the wildlife
//  RedLightPenalty					- Did not stop for a red light
//  BaleDamage
//  GoodDistanceToFire
//  MotorcadeDistance
//  Roadkill
//  BarrelCollision
//  GoodDistanceToSmoke
//  MotorcadePositioning
//  RoughDriving
//  BlindedOtherDrivers
//  GoodsCollision
//  MotorcadeRPM
//  RpmWarning
//  BorrowedFuel
//  InspectionMisjudgement
//  MoveBall
//  Speeding
//  BuildingMaterialCollision
//  InspectionPoints
//  MovePole
//  SpeedingWarning
//  CargoDelivered
//  InspectionWrongOrder
//  MovedToolUnlocked
//  TooCloseToFire
//  ChainInAir
//  LeaveMaze
//  NoSupportLegs
//  TooCloseToSmoke
//  ConeCollisions
//  LeftExerciseArea
//  NotSupportedClaw
//  ToolCollision
//  CurbCollision 					- Drove up on the curb
//  LeftTheTrailerBehind
//  ObstacleCourseTouch
//  UsingParkBrakeWhileDriving
//  DamageToProperty
//  LiftedLoadTooHigh
//  ObstacleMoved
//  VehicleDamage
//  DrivingWithOpenDoors
//  LogsFacingCabin
//  ObstructedTraffic
//  VehicleOffroad
//  DrivingWithSupportLegs
//  LooseCargo
//  PalletMoved
//  VehicleTilt_Flipped
//  DroppedPipe
//  MinorCollision
//  PeopleCollision
//  WheelInAir
//  DumpTruckDamage
//  MissedStation
//  PerfectPlacement
//  WheelOffroad
//  ExcessiveSpeeding
//  MissionsCompleted
//  RanStopSign
type EventData struct {
	BaseData
	Event string
}

// If Type == "ExerciseStart" this is the data send from the simulator
// This indicates that an exercise has just been started.
// ExerciseName is the human readable name of the exercise.
type ExerciseStart struct {
	BaseData
	ExerciseName string
}

// If Type == "ExerciseEnd" this is the data send from the simulator
// This event is sent when an exercise has finished
type ExerciseEnd struct {
	BaseData
}

func (t Timestamp) String() string {
	return fmt.Sprintf("%d:%d:%d.%d", t.Hour, t.Minute, t.Second, t.Millisecond)
}

func (e EventData) String() string {
	return fmt.Sprintf("Time: %v\tEvent: %s", e.Timestamp, e.Event)
}

func (s StreamData) String() string {
	return fmt.Sprintf("Time: %v\tSpeed: %d/%d\tFuelConsumption: %f", s.Timestamp, s.Speed, s.SpeedLimit, s.FuelConsumption)
}

func (s ExerciseStart) String() string {
	return fmt.Sprintf("Time: %v\tExerciseStart: %s", s.Timestamp, s.ExerciseName)
}

func (s ExerciseEnd) String() string {
	return fmt.Sprintf("Time: %v\t ExerciseEnd", s.Timestamp)
}

func main() {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:1534"))
	if err != nil {
		panic(err)
	}

	connection, err := net.DialTCP("tcp", nil, address)
	defer connection.Close()
	if err != nil {
		panic(err)
	}

	var packetData []byte = nil
	braceCount := 0

	for {
		receivedData := make([]byte, 1024)

		// Read data from tcp socket
		receivedCount, err := connection.Read(receivedData)
		if err != nil {
			panic(err)
		}

		if receivedCount > 0 {

			// Count number of brances in received data
			for i := 0; i < receivedCount; i += 1 {
				packetData = append(packetData, receivedData[i])

				if receivedData[i] == byte('{') { // json object start?
					braceCount += 1
				} else if receivedData[i] == byte('}') { // json object end?
					braceCount -= 1
				}

				// If we have a a brace count of zero and some packet data we have received a valid json object.
				if braceCount == 0 && len(packetData) > 0 {
					var base BaseData
					err = json.Unmarshal(packetData, &base)
					if err != nil {
						panic(err)
					}

					if base.Type == "Event" {
						var event EventData
						err = json.Unmarshal(packetData, &event)
						fmt.Println(event.String())
					} else if base.Type == "ExerciseStart" {
						var event ExerciseStart
						err = json.Unmarshal(packetData, &event)
						fmt.Println(event.String())
					} else if base.Type == "ExerciseEnd" {
						var event ExerciseEnd
						err = json.Unmarshal(packetData, &event)
						fmt.Println(event.String())
					} else {
						var stream StreamData
						err = json.Unmarshal(packetData, &stream)
						fmt.Println(stream.String())
					}

					packetData = nil
				}
			}
		}

	}
}
