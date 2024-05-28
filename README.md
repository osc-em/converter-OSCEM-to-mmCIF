# Converter for JSON schema to mmCIF for PDB

This repository implements a file converter JSON <--> PDBx/mmCIF.
The JSON schema is defined in [OSCEM](https://github.com/osc-em/OSCEM_Schemas/) repo and  the PDBx/mmCIF format in ([PDBML](https://mmcif.wwpdb.org/dictionaries/ascii/mmcif_pdbx_v50.dic) Schema v50 ).

The file mapper.tsv is produced by extracting two relevant columns from [conversions in OSCEM](https://github.com/osc-em/OSCEM_Schemas/blob/main/conversions.csv). This file is required to run the Go code.

## To convert JSON into mmCIF:
* specify both Instrument and Sample JSON files ( It is assumed that JSON files pass validation against OSCEM schemas)
* specify the mapper file
* TODO specify the output file
`./converter.go .../OSCEM_Schemas/Instrument/test_data_valid.json .../OSCEM_Schemas/Sample/Sample_valid.json mapper.tsv`


## TBA
* actual creation of mmCIF (now features are extracted, but not written in the output file)
* mmCIF to JSON converter. This, however, is different: EMDB enables download of a very rich mmCIF files with coordinates data. The converter will only create the JSON based on OSCEM schema.