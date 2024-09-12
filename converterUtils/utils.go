// Package converterUtils provides set of simple functions that are required in converter Package
package converterUtils

import "math"

// GetKeys returns a slice of string arrays with keys in map. A key must be a string value.
func GetKeys[K string, V any](m map[string]V) []string {

	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func AssertFloatEqual(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	} else {
		return a == b
	}
}

// PDBxItem type defines attributes of the data item property in PDBx dictionary.
// It contains most important fields, such as name of the item and its parental catrgory,
// type of value it should take on and units, range or allowed values if there are any.
type PDBxItem struct {
	CategoryID string
	Name       string
	Unit       string
	ValueType  string
	RangeMin   string
	RangeMax   string // will be converted only for comparison
	EnumValues []string
}

var UnitsName = map[string]string{
	"e/Å²":  "electrons_angstrom_squared",
	"s":     "seconds",
	"eV":    "electron_volts",
	"kV":    "kilovolts",
	"µm":    "micrometres",
	"mm":    "millimetres",
	"nm":    "nanometers",
	"K":     "kelvins",
	"°":     "degrees",
	"mrad":  "milliradians",
	"Da":    "daltons",
	"mg/ml": "mg_per_ml",
	"%":     "per cent",
}

var PDBxCategoriesOrder = []string{"entry",
	"audit",
	"audit_conform",
	"database",
	"database_2",
	"pdbx_audit_revision_history",
	"pdbx_audit_revision_details",
	"pdbx_audit_revision_group",
	"pdbx_audit_revision_category",
	"pdbx_audit_revision_item",
	"pdbx_audit_support",
	"database_PDB_rev",
	"database_PDB_rev_record",
	"database_PDB_caveat",
	"pdbx_database_PDB_obs_spr",
	"pdbx_database_status",
	"pdbx_database_related",
	"pdbx_database_proc",
	"pdbx_contact_author",
	"audit_contact_author",
	"audit_author",
	"citation",
	"citation_author",
	"citation_editor",
	"database_PDB_remark",
	"entity",
	"entity_keywords",
	"entity_name_com",
	"entity_name_sys",
	"entity_poly",
	"pdbx_entity_nonpoly",
	"entity_poly_seq",
	"entity_src_gen",
	"entity_src_nat",
	"pdbx_entity_src_syn",
	"entity_link",
	"pdbx_entity_branch",
	"pdbx_entity_branch_descriptor",
	"pdbx_entity_branch_link",
	"chem_comp",
	"pdbx_chem_comp_identifier",
	"pdbx_poly_seq_scheme",
	"pdbx_branch_scheme",
	"pdbx_entity_instance_feature",
	"pdbx_nonpoly_scheme",
	"pdbx_unobs_or_zero_occ_atoms",
	"software",
	"computing",
	"cell",
	"symmetry",
	"exptl",
	"exptl_crystal",
	"exptl_crystal_grow",
	"exptl_crystal_grow_comp",
	"diffrn",
	"diffrn_detector",
	"diffrn_radiation",
	"diffrn_radiation_wavelength",
	"diffrn_source",
	"reflns",
	"reflns_shell",
	"refine",
	"refine_analyze",
	"refine_hist",
	"refine_ls_restr",
	"refine_ls_restr_ncs",
	"refine_ls_shell",
	"pdbx_refine",
	"pdbx_xplor_file",
	"struct_ncs_oper",
	"struct_ncs_dom",
	"struct_ncs_dom_lim",
	"struct_ncs_ens",
	"struct_ncs_ens_gen",
	"database_PDB_matrix",
	"struct",
	"struct_keywords",
	"struct_asym",
	"struct_ref",
	"struct_ref_seq",
	"struct_ref_seq_dif",
	"pdbx_struct_assembly",
	"pdbx_struct_assembly_prop",
	"pdbx_struct_assembly_gen",
	"pdbx_struct_assembly_auth_evidence",
	"pdbx_struct_assembly_auth_classification",
	"pdbx_struct_oper_list",
	"struct_biol",
	"struct_biol_gen",
	"struct_biol_view",
	"struct_conf",
	"struct_conf_type",
	"struct_conn",
	"struct_conn_type",
	"pdbx_struct_conn_angle",
	"pdbx_modification_feature",
	"struct_mon_prot_cis",
	"struct_sheet",
	"struct_sheet_order",
	"struct_sheet_range",
	"struct_sheet_hbond",
	"pdbx_struct_sheet_hbond",
	"struct_site",
	"struct_site_gen",
	"pdbx_entry_details",
	"pdbx_validate_close_contact",
	"pdbx_validate_symm_contact",
	"pdbx_validate_rmsd_bond",
	"pdbx_validate_rmsd_angle",
	"pdbx_validate_torsion",
	"pdbx_validate_peptide_omega",
	"pdbx_validate_chiral",
	"pdbx_validate_planes",
	"pdbx_validate_planes_atom",
	"pdbx_validate_main_chain_plane",
	"pdbx_validate_polymer_linkage",
}

var PDBxCategoriesOrderAtom = []string{
	"atom_sites",
	"atom_sites_alt",
	"atom_sites_footnote",
	"atom_type",
	"atom_site",
	"atom_site_anisotrop"}
