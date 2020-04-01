package main

import "database/sql"

func dispatch(db *sql.DB, flags flags) ierrori {

	var ie ierrori

	if flags.product && flags.add {
		ie = addProduct(db)
		if ie != nil {
			return ie
		}
	} else if flags.product && flags.list {
		ie = listProducts(db)
		if ie != nil {
			return ie
		}
	} else if flags.today && flags.add {
		ie = addTodayProduct(db)
		if ie != nil {
			return ie
		}
	} else if flags.norm && flags.add {
		ie = addDailyNorm(db)
		if ie != nil {
			return ie
		}
	} else if flags.product && flags.remove {
		ie = removeProduct(db)
		if ie != nil {
			return ie
		}
	} else if flags.today && flags.remove {
		ie = removeTodayProduct(db)
		if ie != nil {
			return ie
		}
	} else if flags.today && flags.total {
		ie = showTodayTotal(db)
		if ie != nil {
			return ie
		}
	} else {
		return ierror{m: "Unkown flag combination"}
	}

	return nil
}
