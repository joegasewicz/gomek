package libroutey

/*
#include "routey.h"
*/
import "C"

func Router() {
	C.routey()
}
