package main

import "math"

// Constants for moist-air thermodynamics.
const (
	gravity   = 9.80665   // m/s^2
	rdGas     = 287.05    // J/(kg*K), dry air specific gas constant
	cpDryAir  = 1005.7    // J/(kg*K), specific heat of dry air at constant pressure
	latentVap = 2501000.0 // J/kg, latent heat of vaporization
	epsilonWV = 0.622     // Rd/Rv, ratio of gas constants for dry air / water vapor
)

// saturationVaporPressureHPa returns saturation vapor pressure (hPa) for a temperature
// in Celsius, via the Bolton (1980) approximation.
func saturationVaporPressureHPa(tempC float64) float64 {
	return 6.112 * math.Exp(17.67*tempC/(tempC+243.5))
}

// saturationMixingRatio returns saturation mixing ratio (kg/kg) for temperature (C) and
// pressure (hPa).
func saturationMixingRatio(tempC, pressureHPa float64) float64 {
	es := saturationVaporPressureHPa(tempC)
	denom := pressureHPa - es
	if denom <= 0.01 {
		denom = 0.01
	}
	return epsilonWV * es / denom
}

// mixingRatioFromDewpoint returns a parcel's actual mixing ratio (kg/kg): the vapor
// pressure of unsaturated air equals the saturation vapor pressure evaluated AT the
// dewpoint, so this is just saturationMixingRatio with the dewpoint in place of temperature.
func mixingRatioFromDewpoint(dewpointC, pressureHPa float64) float64 {
	return saturationMixingRatio(dewpointC, pressureHPa)
}

// virtualTempK returns virtual temperature (K) for temperature (C) and mixing ratio (kg/kg).
func virtualTempK(tempC, mixingRatio float64) float64 {
	return (tempC + 273.15) * (1.0 + 0.61*mixingRatio)
}

// dryAdiabatTemp returns the temperature (C) of a parcel lifted dry-adiabatically from
// (tempC, pressureHPa) to targetPressureHPa, conserving potential temperature.
func dryAdiabatTemp(tempC, pressureHPa, targetPressureHPa float64) float64 {
	theta := (tempC + 273.15) * math.Pow(1000.0/pressureHPa, 0.286)
	tK := theta * math.Pow(targetPressureHPa/1000.0, 0.286)
	return tK - 273.15
}

// moistAdiabatTemp integrates the pseudo-adiabatic lapse rate (in log-pressure
// coordinates) from (tempC, pressureHPa) to targetPressureHPa using RK2 steps fine
// enough to be numerically stable across the largest gaps in the standard pressure levels.
func moistAdiabatTemp(tempC, pressureHPa, targetPressureHPa float64) float64 {
	lnP0 := math.Log(pressureHPa)
	lnP1 := math.Log(targetPressureHPa)

	dTdlnP := func(t, p float64) float64 {
		tK := t + 273.15
		ws := saturationMixingRatio(t, p)
		num := rdGas*tK + latentVap*ws
		den := cpDryAir + (latentVap*latentVap*ws*epsilonWV)/(rdGas*tK*tK)
		return num / den
	}

	steps := int(math.Ceil(math.Abs(lnP1-lnP0) / 0.01))
	if steps < 1 {
		steps = 1
	}
	dlnP := (lnP1 - lnP0) / float64(steps)

	t := tempC
	lnP := lnP0
	for i := 0; i < steps; i++ {
		p := math.Exp(lnP)
		k1 := dTdlnP(t, p)
		tMid := t + k1*dlnP*0.5
		pMid := math.Exp(lnP + dlnP*0.5)
		k2 := dTdlnP(tMid, pMid)
		t += k2 * dlnP
		lnP += dlnP
	}
	return t
}

// lclBolton computes LCL temperature (K) via Bolton (1980), then LCL pressure (hPa) by
// conserving the parcel's dry potential temperature from the starting level up to the LCL.
func lclBolton(tempC, dewpointC, pressureHPa float64) (lclTempK, lclPressureHPa float64) {
	tK := tempC + 273.15
	tdK := dewpointC + 273.15
	lclTempK = 1.0/(1.0/(tdK-56.0)+math.Log(tK/tdK)/800.0) + 56.0
	theta := tK * math.Pow(1000.0/pressureHPa, 0.286)
	lclPressureHPa = 1000.0 * math.Pow(lclTempK/theta, 1.0/0.286)
	return
}

// interpHeightAtPressure linearly interpolates height (m) at a target pressure (hPa)
// from the sounding's pressure/height pairs (levels ordered surface -> top).
func interpHeightAtPressure(levels []SoundingLevel, targetP float64) float64 {
	if targetP >= levels[0].PressureHPa {
		return levels[0].HeightM
	}
	last := len(levels) - 1
	if targetP <= levels[last].PressureHPa {
		return levels[last].HeightM
	}
	for i := 0; i < last; i++ {
		p0, p1 := levels[i].PressureHPa, levels[i+1].PressureHPa
		if targetP <= p0 && targetP >= p1 {
			if p0 == p1 {
				return levels[i].HeightM
			}
			frac := (p0 - targetP) / (p0 - p1)
			return levels[i].HeightM + frac*(levels[i+1].HeightM-levels[i].HeightM)
		}
	}
	return levels[last].HeightM
}

// interpTempAtHeight linearly interpolates environment temperature (C) at a target
// height (m) AGL-equivalent (actually MSL, matching levels[].HeightM). Returns ok=false
// if the target height falls outside the profile.
func interpTempAtHeight(levels []SoundingLevel, targetH float64) (tempC float64, ok bool) {
	last := len(levels) - 1
	if targetH < levels[0].HeightM || targetH > levels[last].HeightM {
		return 0, false
	}
	for i := 0; i < last; i++ {
		h0, h1 := levels[i].HeightM, levels[i+1].HeightM
		if targetH >= h0 && targetH <= h1 {
			if h1 == h0 {
				return levels[i].TempC, true
			}
			frac := (targetH - h0) / (h1 - h0)
			return levels[i].TempC + frac*(levels[i+1].TempC-levels[i].TempC), true
		}
	}
	return 0, false
}

func computeLapseRate(levels []SoundingLevel, hLow, hHigh float64) float64 {
	tLow, okLow := interpTempAtHeight(levels, hLow)
	tHigh, okHigh := interpTempAtHeight(levels, hHigh)
	if !okLow || !okHigh || hHigh <= hLow {
		return 0
	}
	return (tLow - tHigh) / ((hHigh - hLow) / 1000.0) // °C per km; positive = cooling with height
}

// computePWAT integrates mixing ratio (derived from dewpoint) over pressure across the
// whole profile to get precipitable water, returned in inches.
func computePWAT(levels []SoundingLevel) float64 {
	var totalMm float64
	for i := 0; i < len(levels)-1; i++ {
		w0 := mixingRatioFromDewpoint(levels[i].DewpointC, levels[i].PressureHPa)
		w1 := mixingRatioFromDewpoint(levels[i+1].DewpointC, levels[i+1].PressureHPa)
		wAvg := (w0 + w1) / 2.0
		dP := (levels[i].PressureHPa - levels[i+1].PressureHPa) * 100.0 // hPa -> Pa
		if dP <= 0 {
			continue
		}
		totalMm += wAvg * dP / gravity // kg/m^2 == mm of liquid water
	}
	return totalMm / 25.4
}

// computeParcelIndices runs a surface-based parcel ascent (dry adiabatic below the LCL,
// pseudo-adiabatic above it) against the environment profile to derive CAPE/CIN and the
// associated LCL/LFC/EL, plus PWAT and a couple of standard lapse rates. levels must be
// ordered surface -> top of atmosphere (as built by fetchSounding).
func computeParcelIndices(levels []SoundingLevel) ParcelIndices {
	var idx ParcelIndices
	if len(levels) < 2 {
		return idx
	}

	sfc := levels[0]
	lclTempK, lclP := lclBolton(sfc.TempC, sfc.DewpointC, sfc.PressureHPa)
	idx.LCLPressureHPa = lclP
	idx.LCLHeightM = interpHeightAtPressure(levels, lclP)

	w0 := mixingRatioFromDewpoint(sfc.DewpointC, sfc.PressureHPa) // parcel's mixing ratio, conserved below the LCL

	parcelTempAt := func(p float64) float64 {
		if p >= lclP {
			return dryAdiabatTemp(sfc.TempC, sfc.PressureHPa, p)
		}
		return moistAdiabatTemp(lclTempK-273.15, lclP, p)
	}

	var cape, cin float64
	var lfcP, elP float64

	prevH := sfc.HeightM
	prevTparcel := sfc.TempC
	prevWparcel := w0
	prevTenv := sfc.TempC
	prevWenv := mixingRatioFromDewpoint(sfc.DewpointC, sfc.PressureHPa)

	for i := 1; i < len(levels); i++ {
		lvl := levels[i]
		p := lvl.PressureHPa
		h := lvl.HeightM

		tParcel := parcelTempAt(p)
		var wParcel float64
		if p >= lclP {
			wParcel = w0
		} else {
			wParcel = saturationMixingRatio(tParcel, p)
		}
		tEnv := lvl.TempC
		wEnv := mixingRatioFromDewpoint(lvl.DewpointC, p)

		tvParcelAvg := (virtualTempK(prevTparcel, prevWparcel) + virtualTempK(tParcel, wParcel)) / 2.0
		tvEnvAvg := (virtualTempK(prevTenv, prevWenv) + virtualTempK(tEnv, wEnv)) / 2.0
		dz := h - prevH

		if dz > 0 && tvEnvAvg > 0 {
			buoyancy := gravity * (tvParcelAvg - tvEnvAvg) / tvEnvAvg * dz
			if buoyancy > 0 {
				cape += buoyancy
				if lfcP == 0 {
					lfcP = p
				}
				elP = p
			} else if lfcP == 0 {
				// CIN is the negative area the parcel must overcome to REACH free
				// convection, so only count it below the LFC. A negative dip above
				// an already-reached LFC doesn't count against CIN by convention.
				cin += buoyancy
			}
		}

		prevH = h
		prevTparcel = tParcel
		prevWparcel = wParcel
		prevTenv = tEnv
		prevWenv = wEnv
	}

	idx.CAPEJkg = math.Max(0, cape)
	if lfcP > 0 {
		// Only meaningful as "energy barrier below the LFC" when an LFC actually exists.
		idx.CINJkg = math.Min(0, cin)
		idx.LFCPressureHPa = lfcP
		idx.LFCHeightM = interpHeightAtPressure(levels, lfcP)
	}
	if elP > 0 {
		idx.ELPressureHPa = elP
		idx.ELHeightM = interpHeightAtPressure(levels, elP)
	}

	idx.PrecipitableWaterIn = computePWAT(levels)
	idx.LapseRate0to3kmCPerKm = computeLapseRate(levels, sfc.HeightM, sfc.HeightM+3000)
	idx.LapseRate3to6kmCPerKm = computeLapseRate(levels, sfc.HeightM+3000, sfc.HeightM+6000)
	idx.LapseRate700to500CPerKm = computeLapseRate(levels, interpHeightAtPressure(levels, 700), interpHeightAtPressure(levels, 500))

	return idx
}
