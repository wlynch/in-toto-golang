package cmd

import (
	"fmt"
	"path/filepath"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/spf13/cobra"
)

var (
	stepName       string
	runDir         string
	materialsPaths []string
	productsPaths  []string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes the passed command and records paths and hashes of 'materials'",
	Long: `Executes the passed command and records paths and hashes of 'materials' (i.e.
files before command execution) and 'products' (i.e. files after command
execution) and stores them together with other information (executed command,
return value, stdout, stderr, ...) to a link metadata file, which is signed
with the passed key.  Returns nonzero value on failure and zero otherwise.`,
	Args:    cobra.MinimumNArgs(1),
	PreRunE: getKeyCert,
	RunE:    run,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(
		&stepName,
		"name",
		"n",
		"",
		`Name used to associate the resulting link metadata
with the corresponding step defined in an in-toto layout.`,
	)

	runCmd.Flags().StringVarP(
		&runDir,
		"run-dir",
		"r",
		"",
		`runDir specifies the working directory of the command.
If runDir is the empty string, the command will run in the
calling process's current directory. The runDir directory must
exist, be writable, and not be a symlink.`,
	)

	runCmd.Flags().StringVarP(
		&keyPath,
		"key",
		"k",
		"",
		`Path to a PEM formatted private key file used to sign
the resulting link metadata.`,
	)

	runCmd.Flags().StringVarP(
		&certPath,
		"cert",
		"c",
		"",
		`Path to a PEM formatted certificate that corresponds with
the provided key.`,
	)

	runCmd.Flags().StringArrayVarP(
		&materialsPaths,
		"materials",
		"m",
		[]string{},
		`Paths to files or directories, whose paths and hashes
are stored in the resulting link metadata before the
command is executed. Symlinks are followed.`,
	)

	runCmd.Flags().StringArrayVarP(
		&productsPaths,
		"products",
		"p",
		[]string{},
		`Paths to files or directories, whose paths and hashes
are stored in the resulting link metadata after the
command is executed. Symlinks are followed.`,
	)

	runCmd.Flags().StringVarP(
		&outDir,
		"metadata-directory",
		"d",
		"./",
		`Directory to store link metadata`,
	)

	runCmd.Flags().StringArrayVarP(
		&lStripPaths,
		"lstrip-paths",
		"l",
		[]string{},
		`Path prefixes used to left-strip artifact paths before storing
them to the resulting link metadata. If multiple prefixes
are specified, only a single prefix can match the path of
any artifact and that is then left-stripped. All prefixes
are checked to ensure none of them are a left substring
of another.`,
	)

	runCmd.Flags().StringArrayVarP(
		&exclude,
		"exclude",
		"e",
		[]string{},
		`Path patterns to match paths that should not be recorded as 0
‘materials’ or ‘products’. Passed patterns override patterns defined
in environment variables or config files. See Config docs for details.`,
	)

	runCmd.MarkFlagRequired("name")

	runCmd.Flags().BoolVar(
		&lineNormalization,
		"normalize-line-endings",
		false,
		`Enable line normalization in order to support different
operating systems. It is done by replacing all line separators
with a new line character.`,
	)

	runCmd.Flags().StringVar(
		&spiffeUDS,
		"spiffe-workload-api-path",
		"",
		"UDS path for SPIFFE workload API",
	)

}

func run(cmd *cobra.Command, args []string) error {
	block, err := intoto.InTotoRun(stepName, runDir, materialsPaths, productsPaths, args, key, []string{"sha256"}, exclude, lStripPaths, lineNormalization)
	if err != nil {
		return fmt.Errorf("failed to create link metadata: %w", err)
	}

	linkName := fmt.Sprintf(intoto.LinkNameFormat, block.Signed.(intoto.Link).Name, key.KeyID)

	linkPath := filepath.Join(outDir, linkName)
	err = block.Dump(linkPath)
	if err != nil {
		return fmt.Errorf("failed to write link metadata to %s: %w", linkPath, err)
	}

	return nil
}
