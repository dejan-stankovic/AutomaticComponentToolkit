/*++

Copyright (C) 2018 Automatic Component Toolkit Developers

All rights reserved.

Abstract: This is the class declaration of CLibPrimesFactorizationCalculator

*/


#ifndef __LIBPRIMES_LIBPRIMESFACTORIZATIONCALCULATOR
#define __LIBPRIMES_LIBPRIMESFACTORIZATIONCALCULATOR

#include "libprimes_interfaces.hpp"

// Parent classes
#include "libprimes_calculator.hpp"
#pragma warning( push)
#pragma warning( disable : 4250)

// Include custom headers here.
#include <vector>

namespace LibPrimes {


/*************************************************************************************************************************
 Class declaration of CLibPrimesFactorizationCalculator 
**************************************************************************************************************************/

class CLibPrimesFactorizationCalculator : public virtual ILibPrimesFactorizationCalculator, public virtual CLibPrimesCalculator {
private:

	std::vector<sLibPrimesPrimeFactor> primeFactors;

protected:

	/**
	* Put protected members here.
	*/

public:

	/**
	* Put additional public members here. They will not be visible in the external API.
	*/

	void Calculate();

	/**
	* Public member functions to implement.
	*/

	void GetPrimeFactors (LibPrimes_uint64 nPrimeFactorsBufferSize, LibPrimes_uint64 * pPrimeFactorsNeededCount, sLibPrimesPrimeFactor * pPrimeFactorsBuffer);

	bool CheckPrimeFactors (const LibPrimes_uint64 nPrimeFactorsBufferSize, const sLibPrimesPrimeFactor * pPrimeFactorsBuffer);
};

}

#pragma warning( pop )
#endif // __LIBPRIMES_LIBPRIMESFACTORIZATIONCALCULATOR
