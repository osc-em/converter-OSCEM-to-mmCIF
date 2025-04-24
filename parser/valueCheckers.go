package parser

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

func validateDateIsRFC3339(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Printf(
			"Date seems to be in the wrong format! We expect RFC3339 format, provided date %s is not.", date)
		return ""
	}
	return t.Format(time.DateOnly)
}

func validateRange(value string, dataItem converterUtils.PDBxItem, unitOSCEM string, nameOSCEM string) (bool, error) {
	rMin := dataItem.RangeMin
	rMax := dataItem.RangeMax
	unitPDBx := dataItem.Unit
	namePDBx := dataItem.CategoryID + "." + dataItem.Name
	var unitsSame bool
	var errorMessage string
	var unitsError error
	if unitOSCEM == "" && unitPDBx == "" {
		// both OSCEM and PDBx have no units definition
		unitsSame = true
	} else if unitOSCEM == "" && unitPDBx != "" {
		errorMessage = fmt.Sprintf(
			"No units defined for %s in OSCEM! Analogous property %s in PDBx has %s units. Value will still be used in mmCIF file!",
			nameOSCEM, namePDBx, unitPDBx)
		unitsSame = false
		unitsError = errors.New(errorMessage)
		return true, unitsError
	} else if unitOSCEM != "" && unitPDBx == "" {
		errorMessage = fmt.Sprintf(
			"No units defined for %s in PDBx! Analogous property %s in OSCEM has %s units. Value will still be used in mmCIF file!",
			namePDBx, nameOSCEM, unitOSCEM)
		unitsSame = false
		unitsError = errors.New(errorMessage)
		return true, unitsError
	} else {
		explicitUnitOSCEM, ok := converterUtils.UnitsName[unitOSCEM]
		if !ok {
			errorMessage = fmt.Sprintf(
				"No explicit unit name is specified for property %s in OSCEM, only a short name %s. Value will still be used in mmCIF file!",
				nameOSCEM, unitOSCEM)
			unitsSame = false
			unitsError = errors.New(errorMessage)
			return true, unitsError
		} else {
			unitsError = nil
		}
		unitsSame = explicitUnitOSCEM == unitPDBx
	}
	// FIXME: when units are settled, do conversion when possible!
	if unitsSame {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			errorMessage = fmt.Sprintf("JSON value %s not numeric, but supposed to be", value)
			return false, errors.New(errorMessage)
		}
		rMin, err := strconv.ParseFloat(rMin, 64)
		if err != nil {
			rMin = math.NaN()
		}
		rMax, err := strconv.ParseFloat(rMax, 64)
		if err != nil {
			rMax = math.NaN()
		}
		if math.IsNaN(rMin) && math.IsNaN(float64(rMax)) {
			return true, unitsError
		} else if math.IsNaN(rMin) {
			return float64(v) <= rMax, unitsError
		} else if math.IsNaN(rMax) {
			return float64(v) >= rMin, unitsError
		} else {
			return float64(v) >= rMin && float64(v) <= rMax, unitsError
		}
	} else {
		errorMessage = fmt.Sprintf(
			"Units for analogous properties %s in OSCEM and %s in PDBx  don't match!"+
				" Implement a converter from %s in OSCEM to %s expected by PDBx. Value will still be used in mmCIF file!",
			nameOSCEM, namePDBx, unitOSCEM, unitPDBx)
		unitsError = errors.New(errorMessage)
		return true, unitsError
	}

}
func validateEnum(value string, dataItem converterUtils.PDBxItem) string {
	enumFromEMDB := dataItem.EnumValues
	enumFromPDBx := dataItem.PDBxEnumValues
	namePDBx := dataItem.CategoryID + "." + dataItem.Name
	if value == "true" {
		return "YES"
	} else if value == "false" {
		return "NO"
	}

	if namePDBx == "_em_imaging.microscope_model" {
		reTitan := regexp.MustCompile(`(?i)titan`)
		if reTitan.MatchString(value) {
			return "TFS KRIOS"
		}
	} else if namePDBx == "_em_imaging.mode" {
		if value == "BrightField" {
			return "BRIGHT FIELD"
		}
	} else if namePDBx == "_em_imaging.electron_source" {
		if value == "FieldEmission" {
			return "FIELD EMISSION GUN"
		}
	}
	for i := range enumFromEMDB {
		if strings.EqualFold(enumFromEMDB[i], value) {
			value = enumFromEMDB[i]
			return value
		}
	}
	// scan through both enums
	for i := range enumFromPDBx {
		if strings.EqualFold(enumFromPDBx[i], value) {
			value = enumFromPDBx[i]
			return value
		}
	}
	if namePDBx == "_em_imaging.illumination_mode" {
		reFloodBeam := regexp.MustCompile(`(?i)parallel`)
		if reFloodBeam.MatchString(value) {
			return "FLOOD BEAM"
		}
	}
	if namePDBx == "_pdbx_contact_author.role" {
		reRole := regexp.MustCompile(`(?i)(principal investigator|group leader|pi)`)
		if reRole.MatchString(value) {
			return "principal investigator/group leader"
		}
	}
	// add additional matching mechanism for grid material by chemical element name/ regular expression
	if namePDBx == "_em_sample_support.grid_material" {

		reGraphene := regexp.MustCompile(`(?i)graphene`)
		reSilicon := regexp.MustCompile(`(?i)silicon`)
		if reGraphene.MatchString(value) {
			return "GRAPHENE OXIDE"
		} else if reSilicon.MatchString(value) {
			return "SILICON NITRIDE"
		}
		switch value {
		case "Cu":
			return "COPPER"
		case "Cu/Pd":
			return "COPPER/PALLADIUM"
		case "Cu/Rh":
			return "COPPER/RHODIUM"
		case "Au":
			return "GOLD"
		case "Ni":
			return "NICKEL"
		case "Ni/Ti":
			return "NICKEL/TITANIUM"
		case "Pt":
			return "PLATINUM"
		case "W":
			return "TUNGSTEN"
		case "Ti":
			return "TITANIUM"
		case "Mo":
			return "MOLYBDENUM"
		}

	} else if namePDBx == "_em_image_recording.film_or_detector_model" {
		// add additional matching mechanism for "Falcon" detector model by regular expression; other don't seem feasible

		reFalconI := regexp.MustCompile(`(?i)falcon[\s_]*?(1|I)`)
		reFalconII := regexp.MustCompile(`(?i)falcon[\s_]*?(2|II)`)
		reFalconIII := regexp.MustCompile(`(?i)falcon[\s_]*?(3|III)`)
		reFalconIV := regexp.MustCompile(`(?i)falcon[\s_]*?(4|IV)`)
		switch {
		case reFalconIV.MatchString(value):
			return "FEI FALCON IV (4k x 4k)"
		case reFalconIII.MatchString(value):
			return "FEI FALCON III (4k x 4k)"
		case reFalconII.MatchString(value):
			return "FEI FALCON II (4k x 4k)"
		case reFalconI.MatchString(value):
			return "FEI FALCON I (4k x 4k)"
		}
	}

	log.Printf("value %v is not in enum %s!", value, namePDBx)
	// if not found in enum list and it's a funding organisation, put a certain string
	if namePDBx == "_pdbx_audit_support.funding_organization" {
		return "Other government"
	}
	// // if no match and enum contains option for OTHER, choose it
	// if converterUtils.SliceContains(enumFromEMDB, "OTHER") || converterUtils.SliceContains(enumFromPDBx, "OTHER"){
	// 	return "OTHER"
	// }

	return strings.ToUpper(value)
}

func checkValue(dataItem converterUtils.PDBxItem, value string, jsonKey string, unitsOSCEM string) string {

	//now based on the found struct implement range matching, units matching and enum matching
	if dataItem.ValueType == "int" || dataItem.ValueType == "float" {
		namePDBx := dataItem.CategoryID + "." + dataItem.Name
		// in OSCEM defocus is negative value and overfocus is positive. In PDBx it's vice versa if string starts with  minus, ut it off, otherwise add a - prefix
		if namePDBx == "_em_imaging.nominal_defocus_min" ||
			namePDBx == "_em_imaging.calibrated_defocus_min" ||
			namePDBx == "_em_imaging.nominal_defocus_max" ||
			namePDBx == "_em_imaging.calibrated_defocus_max" {
			// change the sign negative to positive and vice versa
			if value[0] == 45 {
				value = value[1:]
			} else {
				value = "-" + value
			}
		}
		validatedRange, err := validateRange(value, dataItem, unitsOSCEM, jsonKey)
		if err != nil {
			errorNumeric := fmt.Sprintf("JSON value %s not numeric, but supposed to be", value)
			if err.Error() == errorNumeric {
				return "?"
			} else {
				// FIXME when units conversion is implemented, handle this error
				log.Println(err.Error())
			}
		}
		if !validatedRange {
			log.Printf(
				"Value %s of property %s is not in range of [ %s, %s ]!\n", value, jsonKey, dataItem.RangeMin, dataItem.RangeMax)
		}
	} else if dataItem.ValueType == "yyyy-mm-dd" {
		value = validateDateIsRFC3339(value)
	} else if len(dataItem.EnumValues) > 0 || len(dataItem.PDBxEnumValues) > 0 {
		value = validateEnum(value, dataItem)
	}

	if strings.Contains(value, " ") {
		value = fmt.Sprintf("'%s' ", value) // if name contains whitespaces enclose it in single quotes
	} else {
		value = fmt.Sprintf("%s ", value) // take value as is
	}
	return value
}
