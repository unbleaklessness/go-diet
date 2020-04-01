package main

import "database/sql"

func dispatch(db *sql.DB, flags flags) (ie *ierror) {

	if flags.product && flags.add {
		ie = addProduct(db)
		if ie != nil {
			return
		}
	} else if flags.product && flags.list {
		ie = listProducts(db)
		if ie != nil {
			return
		}
	} else if flags.today && flags.add {
		ie = addTodayProduct(db)
		if ie != nil {
			return
		}
	} else if flags.norm && flags.add {
		ie = addDailyNorm(db)
		if ie != nil {
			return
		}
	} else {
		ie = &ierror{m: "Unkown flag combination"}
		return
	}

	return
}
