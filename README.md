# Converter for JSON schema to mmCIF for PDB

This repository implements a file converterfrom OSC-EM JSON to PDBx/mmCIF.
The JSON schema is defined in [OSCEM](https://github.com/osc-em/OSCEM_Schemas/) and  the PDBx/mmCIF format in ([PDBML](https://mmcif.wwpdb.org/dictionaries/ascii/mmcif_pdbx_v50.dic) Schema v50 ).


## Running the converter and required inputs:
** Option 1 ** 
You can download the executable from the releases and use it in terminal with the following flags: 
* with `--json` specify path to json file that contains metadata
* with `--dic` specify path to the PDBx/mmCIF dictionary file
* with `--conversions` specify path to [conversions table](https://github.com/openem/LS_Metadata_reader/). This table includes correspondance in names between OSC-EM and PDBx
* with `--level` specify the json element name that contains metadata entries. For SciCat that is usually "scientificMetadata"
* with `--append` specify if the metadata should be added to existing mmCIF to later deposit it in PDB
* with `--mmCIFfile` specify the path to existing mmCIF file. Throws an error if --append is false and --mmCIFfile is not specified
* with `--output` specify the file to write the newly created mmCIF with metadata entries

** Option 2 **
If using in another GO application, you can install this package with 
```go get github.com/osc-em/converter-OSCEM-to-mmCIF 
```
and then run it with some of the 

```import parser "github.com/osc-em/converter-OSCEM-to-mmCIF/parser"
    ...
    parser.EMDBconvert(scientificMetadata, metadataLevelName, conversionsCsvPath, mmCIFdictPath, outputPath) // or
    parser.PDBconvertFromPath(scientificMetadata, metadataLevelName, conversionsCsvPath, mmCIFdictPath, outputPath, mmCIFInputPath)
```
The first argument in both functions is a scientificMetadata of type `map[string]any` that expects an unmarshalled JSON. 

Here no checks against [OSCEM schema](https://github.com/osc-em/OSCEM_Schemas) are implemented, it is assumed metadata validates against this schema.  Validation might still be added in the next releases.



## Checks against mmCIF
Converter explicitly parser through the PDBx definitions to extract as much data as possible. This allows for
* administrative categories sorted within mmCIF ( such as author information, grant, etc)
* em-related categories are sorted randomly, as there is no definitive sorting in PDB team as well
* file ends with information on atoms
* units in OSCEM definition are comared to PDBx ( converter for units will be implemented)
* numeric values are checked to be within a range allowed by PDBs
* values  are checked to be within a list of attributes allowed by PDBx. This is additionally enhanced to match via regular expressions or ceratin logic. 