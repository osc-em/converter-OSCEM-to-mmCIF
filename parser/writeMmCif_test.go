package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/osc-em/converter-OSCEM-to-mmCIF/converterUtils"
)

func TestGetKeyByValue(t *testing.T) {
	var tests = []struct {
		name           string
		value          string
		dictionary     map[string]string
		expectedResult string
		expectedError  string
	}{
		{"value exists in the map", "world", map[string]string{"hello": "world"}, "hello", ""},
		{"value does not exist in a map", "hello", map[string]string{"hello": "world"}, "", "value hello is not in the conversion table"},
	}

	for _, test := range tests {

		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := getKeyByValue(test.value, test.dictionary)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedResult {
				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
			}
		})
	}
}

// func TestSliceContains(t *testing.T) {
// 	var testSlice = []struct {
// 		name           string
// 		slice          []string
// 		element        string
// 		expectedResult bool
// 	}{
// 		{"element in slice", []string{"hello", "world"}, "hello", true},
// 		{"element not in slice", []string{"hello", "world"}, "foo", false},
// 	}

// 	for _, test := range testSlice {

// 		testname := fmt.Sprintf("%v", test.name)
// 		t.Run(testname, func(t *testing.T) {
// 			gotValue := sliceContains(test.slice, test.element)

// 			if gotValue != test.expectedResult {
// 				t.Errorf("got %v, want %v", gotValue, test.expectedResult)
// 			}
// 		})
// 	}
// }

func TestToEMDB(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		expectedText  string
		expectedError string
	}{
		{
			"a data category with no arrays in JSON and valid values",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"data_myID\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\n",
			"",
		},
		{
			"a data category with array in JSON and valid values",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1", "1.5"}, "foo.goo": {"3.14157", "2.4"}, "foo.foo": {"2.3", "1.1"}},
			map[string][]string{"foo.boo": {"s", "s"}, "foo.goo": {"u2", "u2"}},
			"data_myID\n#\nloop_\ncat1.name1\ncat1.name2\ncat1.name22\n1 3.14157 2.3 \n1.5 2.4 1.1 \n#\n",
			"",
		},
		{
			"function call to date and that a whole PDBx category is skipped on no JSON values are present for it ",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.roo": {"2024-03-08T19:11:59+01:00"}},
			map[string][]string{},
			"data_myID\n#\ncat1.name1     2024-03-08 \n#\n",
			"",
		},
		{
			"a non-numeric value for a range",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "60", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.goo": {"ten"}},
			map[string][]string{"foo.goo": {"s"}},
			"data_myID\n#\ncat1.name2     ?\n#\n",
			"",
		},
		{
			"function call to enum checkers and single quotes are added when space in the value",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "foo.doo": "cat2.name1"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello world", "foo boo"}},
				},
				"cat2": {
					{CategoryID: "cat2", Name: "name1", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"foo.foo": {"hello world"}},
			map[string][]string{},
			"data_myID\n#\ncat1.name3     'hello world' \n#\n",
			"",
		},
		{
			"sign switch for defocus",
			map[string]string{"foo.roo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name3", "metadata.defocus.min": "_em_imaging.nominal_defocus_min", "metadata.defocus.max": "_em_imaging.nominal_defocus_max"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", ValueType: "yyyy-mm-dd"},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello world", "foo boo"}},
				},
				"_em_imaging": {
					{CategoryID: "_em_imaging", Name: "nominal_defocus_min", ValueType: "float", RangeMin: "0", RangeMax: "1500", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "_em_imaging", Name: "nominal_defocus_max", ValueType: "float", RangeMin: "0", RangeMax: "1500", EnumValues: []string{}, PDBxEnumValues: []string{}},
				},
			},
			map[string][]string{"metadata.defocus.min": {"-1100"}, "metadata.defocus.max": {"300"}},
			map[string][]string{},
			"data_myID\n#\n_em_imaging.nominal_defocus_min     1100 \n_em_imaging.nominal_defocus_max     -300 \n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := CreteMetadataCif(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}

func TestToPDB(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		pathExisting  string
		expectedText  string
		expectedError string
	}{
		{
			"no input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"someFile.cif",
			"",
			"mmCIF file someFile.cif does not exist!",
		},
		{
			"valid input mmCIF",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			gotValue, gotError := SupplementCoordinatesFromPath(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM, test.pathExisting)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}

// all cif provided here are valid.
func TestToPDB2(t *testing.T) {
	var testCases = []struct {
		name          string
		namesMap      map[string]string
		PDBxItems     map[string][]converterUtils.PDBxItem
		jsonValues    map[string][]string
		unitsOSCEM    map[string][]string
		pathExisting  string
		expectedText  string
		expectedError string
	}{
		{
			"Categories complement each other",
			map[string]string{"foo.boo": "cat1.name1", "foo.goo": "cat1.name2", "foo.foo": "cat1.name22", "foo.doo": "cat1.name3"},
			map[string][]converterUtils.PDBxItem{
				"cat1": {
					{CategoryID: "cat1", Name: "name1", Unit: "seconds", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name2", Unit: "u3", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name22", ValueType: "float", RangeMin: "0", RangeMax: "3.5", EnumValues: []string{}, PDBxEnumValues: []string{}},
					{CategoryID: "cat1", Name: "name3", EnumValues: []string{"hello", "world"}},
				},
			},
			map[string][]string{"foo.boo": {"1"}, "foo.goo": {"3.14157"}, "foo.foo": {"2.3"}},
			map[string][]string{"foo.boo": {"s"}, "foo.goo": {"u2"}},
			"testData/example.cif",
			"data_K3DAK4\n#\nloop_\n_citation.id\n_citation.title\n_citation.journal_abbrev\n_citation.journal_volume\n_citation.page_first\n_citation.page_last\n_citation.year\n_citation.journal_id_ASTM\n_citation.journal_id_ISSN\n_citation.journal_id_CSD\nphenix.real_space_refine 'Real-space refinement in PHENIX for cryo-EM and crystallography' 'Acta Crystallogr., Sect. D: Biol. Crystallogr.' 74 531 544 2018 ABCRE6 0907-4449 0766\n#\nloop_\n_chem_comp.id\nALA\nARG\nASN\nASP\nCYS\nGLN\nGLU\nGLY\nHIS\nILE\nLEU\nLYS\nMET\nPHE\nPRO\nSER\nTHR\nTRP\nTYR\nVAL\n#\nloop_\n_software.pdbx_ordinal\n_software.name\n_software.version\n_software.type\n_software.contact_author\n_software.contact_author_email\n_software.location\n_software.classification\n_software.citation_id\n_software.language\n1 phenix.real_space_refine 1.20rc4_4425 program 'Pavel Afonine' pafonine@lbl.gov https://www.phenix-online.org/ refinement phenix.real_space_refine Python/C++\n1 Phenix 1.20rc4_4425 program 'Paul D. Adams' pdadams@lbl.gov https://www.phenix-online.org/ refinement phenix Python/C++\n#\nloop_\n_space_group_symop.id\n_space_group_symop.operation_xyz\n1 x,y,z\n#\ncat1.name1      1 \ncat1.name2      3.14157 \ncat1.name22     2.3 \n#\nloop_\n_atom_site.group_PDB\n_atom_site.id\n_atom_site.label_atom_id\n_atom_site.label_alt_id\n_atom_site.label_comp_id\n_atom_site.auth_asym_id\n_atom_site.auth_seq_id\n_atom_site.pdbx_PDB_ins_code\n_atom_site.Cartn_x\n_atom_site.Cartn_y\n_atom_site.Cartn_z\n_atom_site.occupancy\n_atom_site.B_iso_or_equiv\n_atom_site.type_symbol\n_atom_site.pdbx_formal_charge\n_atom_site.label_asym_id\n_atom_site.label_entity_id\n_atom_site.label_seq_id\n_atom_site.pdbx_PDB_model_num\nATOM 1 N . SER B 535 ? 270.43781 345.22081 281.42585 1.000 465.59921 N ? A ? 1 1\nATOM 2 CA . SER B 535 ? 270.07764 346.63283 281.41226 1.000 465.59921 C ? A ? 1 1\nATOM 3 C . SER B 535 ? 268.70572 346.84369 282.04532 1.000 465.59921 C ? A ? 1 1\nATOM 4 O . SER B 535 ? 268.12252 345.92083 282.61134 1.000 465.59921 O ? A ? 1 1\nATOM 5 CB . SER B 535 ? 270.08494 347.17423 279.98114 1.000 465.59921 C ? A ? 1 1\nATOM 6 OG . SER B 535 ? 271.31962 346.90634 279.34048 1.000 465.59921 O ? A ? 1 1\nATOM 7 N . VAL B 536 ? 268.19595 348.07334 281.94291 1.000 465.59921 N ? A ? 2 1\nATOM 8 CA . VAL B 536 ? 266.88271 348.37783 282.50166 1.000 465.59921 C ? A ? 2 1\nATOM 9 C . VAL B 536 ? 265.76647 347.61528 281.79180 1.000 465.59921 C ? A ? 2 1\nATOM 10 O . VAL B 536 ? 265.89177 347.21437 280.62991 1.000 465.59921 O ? A ? 2 1\nATOM 11 CB . VAL B 536 ? 266.60814 349.89588 282.55950 1.000 465.59921 C ? A ? 2 1\nATOM 12 CG1 . VAL B 536 ? 266.47176 350.50312 281.15572 1.000 465.59921 C ? A ? 2 1\nATOM 13 CG2 . VAL B 536 ? 265.37232 350.19041 283.40137 1.000 465.59921 C ? A ? 2 1\nATOM 14 N . VAL B 537 ? 264.66075 347.41139 282.50950 1.000 465.59921 N ? A ? 3 1\nATOM 15 CA . VAL B 537 ? 263.53325 346.68375 281.94275 1.000 465.59921 C ? A ? 3 1\nATOM 16 C . VAL B 537 ? 263.01065 347.41204 280.70782 1.000 465.59921 C ? A ? 3 1\nATOM 17 O . VAL B 537 ? 263.07692 348.64465 280.60360 1.000 465.59921 O ? A ? 3 1\nATOM 18 CB . VAL B 537 ? 262.41900 346.51315 282.98883 1.000 465.59921 C ? A ? 3 1\nATOM 19 CG1 . VAL B 537 ? 261.74793 347.85051 283.27814 1.000 465.59921 C ? A ? 3 1\nATOM 20 CG2 . VAL B 537 ? 261.39935 345.49031 282.51421 1.000 465.59921 C ? A ? 3 1\n#\nloop_\n_atom_site_anisotrop.id\n_atom_site_anisotrop.pdbx_auth_atom_id\n_atom_site_anisotrop.pdbx_label_alt_id\n_atom_site_anisotrop.pdbx_auth_comp_id\n_atom_site_anisotrop.pdbx_auth_asym_id\n_atom_site_anisotrop.pdbx_auth_seq_id\n_atom_site_anisotrop.pdbx_PDB_ins_code\n_atom_site_anisotrop.U[1][1]\n_atom_site_anisotrop.U[2][2]\n_atom_site_anisotrop.U[3][3]\n_atom_site_anisotrop.U[1][2]\n_atom_site_anisotrop.U[1][3]\n_atom_site_anisotrop.U[2][3]\n1 N . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n2 CA . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n3 C . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n4 O . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n5 CB . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n6 OG . SER B 535 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n7 N . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n8 CA . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n9 C . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n10 O . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n11 CB . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n12 CG1 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n13 CG2 . VAL B 536 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n14 N . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n15 CA . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n16 C . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n17 O . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n18 CB . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n19 CG1 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n20 CG2 . VAL B 537 ? 5.89688 5.89688 5.89688 -0.00000 -0.00000 0.00000\n#\n#\n",
			"",
		},
		{
			"Category and value present in both mmCIF and metadata but different number of instances should take the whole category from mmCIF",
			map[string]string{"mydata1": "_pdbx_nonpoly_scheme.mon_id", "mydata2": "_pdbx_nonpoly_scheme.auth_seq_num", "mydata3": "_pdbx_nonpoly_scheme.pdb_mon_id"},
			map[string][]converterUtils.PDBxItem{
				"_pdbx_nonpoly_scheme": {
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "mon_id"},
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "auth_seq_num"},
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "pdb_mon_id"},
				},
			},
			map[string][]string{"mydata1": {"CU"}, "mydata2": {"218"}, "mydata3": {"CU"}},
			map[string][]string{},
			"testData/exampleBothLoop.cif",
			"data_K3DAK4\n#\nloop_\n_pdbx_nonpoly_scheme.asym_id\n_pdbx_nonpoly_scheme.entity_id\n_pdbx_nonpoly_scheme.mon_id\n_pdbx_nonpoly_scheme.ndb_seq_num\n_pdbx_nonpoly_scheme.pdb_seq_num\n_pdbx_nonpoly_scheme.auth_mon_id\n_pdbx_nonpoly_scheme.pdb_strand_id\n_pdbx_nonpoly_scheme.pdb_ins_code\nB 2 FE 1 201 'FE a' A .\nC 3 ZN 1 202 'ZN b' A .\n#\n",
			"",
		},
		{
			"Category and value present in both mmCIF and metadata and both have one instance, fields should complement each other",
			map[string]string{"mydata1": "_entity_poly.type", "mydata2": "_entity_poly.pdbx_seq_one_letter_code", "mydata3": "_entity_poly.nstd_monomer"},
			map[string][]converterUtils.PDBxItem{
				"_entity_poly": {
					{CategoryID: "_entity_poly", Name: "type"},
					{CategoryID: "_entity_poly", Name: "pdbx_seq_one_letter_code"},
					{CategoryID: "_entity_poly", Name: "nstd_monomer", EnumValues: []string{"yes", "no"}},
				},
			},
			map[string][]string{"mydata1": {"mytype"}, "mydata2": {"mysequence"}, "mydata3": {"yes"}},
			map[string][]string{},
			"testData/exampleBoth.cif",
			"data_K3DAK4\n#\n_entity_poly.type                                      POLYPEPTIDE \n_entity_poly.pdbx_seq_one_letter_code                  'PQPQ HHLLRPRRRK RPHSIPTPIL IFRSP' \n_entity_poly.entity_id                                 1 \n_entity_poly.nstd_linkage                              no \n_entity_poly.nstd_monomer                              no \n_entity_poly.pdbx_strand_id                            A \n_entity_poly.pdbx_target_identifier                    ? \n#\n#\n",
			"",
		},
		{
			"Category and value present in both mmCIF and metadata and both have the same number of instances(>1), fields should complement each other within loop",
			map[string]string{"mydata1": "_pdbx_nonpoly_scheme.mon_id", "mydata2": "_pdbx_nonpoly_scheme.auth_seq_num", "mydata3": "_pdbx_nonpoly_scheme.pdb_mon_id", "mydata4": "_pdbx_nonpoly_scheme.pdb_ins_code"},
			map[string][]converterUtils.PDBxItem{
				"_pdbx_nonpoly_scheme": {
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "mon_id"},
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "auth_seq_num"},
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "pdb_mon_id"},
					{CategoryID: "_pdbx_nonpoly_scheme", Name: "pdb_ins_code"},
				},
			},
			map[string][]string{"mydata1": {"CU", "ZN"}, "mydata2": {"218", "202"}, "mydata3": {"CU", "ZN"}, "mydata4": {"hello", "world"}},
			map[string][]string{},
			"testData/exampleBothLoop.cif",
			"data_K3DAK4\n#\nloop_\n_pdbx_nonpoly_scheme.asym_id\n_pdbx_nonpoly_scheme.entity_id\n_pdbx_nonpoly_scheme.mon_id\n_pdbx_nonpoly_scheme.ndb_seq_num\n_pdbx_nonpoly_scheme.pdb_seq_num\n_pdbx_nonpoly_scheme.auth_mon_id\n_pdbx_nonpoly_scheme.pdb_strand_id\n_pdbx_nonpoly_scheme.pdb_ins_code\nB 2 FE 1 201 'FE a' A . \nC 3 ZN 1 202 'ZN b' A . \n#\n#\n",
			"",
		},
	}

	for _, test := range testCases {
		testname := fmt.Sprintf("%v", test.name)
		t.Run(testname, func(t *testing.T) {
			dictFile, _ := os.Open(test.pathExisting)
			defer dictFile.Close()
			gotValue, gotError := SupplementCoordinates(test.namesMap, test.PDBxItems, test.jsonValues, test.unitsOSCEM, dictFile)

			if gotError != nil {
				if gotError.Error() != test.expectedError {
					t.Errorf("got error %v, wanted %v", gotError.Error(), test.expectedError)
				}
			}
			if gotValue != test.expectedText {
				// for j := range test.expectedText {
				// 	if gotValue[j] != test.expectedText[j] {
				// 		fmt.Printf("position %v, got:%s, want:%s\n", j, string(gotValue[j]), string(test.expectedText[j]))
				// 	}
				// }

				t.Errorf("got:\n%v, want:\n%v", gotValue, test.expectedText)

			}
		})
	}
}
