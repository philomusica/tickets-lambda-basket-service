package main

func CalculateBalance(numOfFullPrice uint8, fullPriceCost float32, numOfConcessions uint8, concessionCost float32) (total float32) {
	total = float32(numOfConcessions) * fullPriceCost + float32(numOfConcessions) * concessionCost
	return total
}
func main() {
}
