package program

// UI Layout Constants

const (
	// Terminal minimum dimensions
	minWidth  = 70
	minHeight = 20

	// Title box dimensions
	titleBoxHeight = 5

	// Table column widths
	tableDateWidth          = 21
	tableCategoryWidth      = 12
	tableAmountWidth        = 12
	tableAmountWidthSummary = 15
	tableCountWidth         = 5
	tableMinDescWidth       = 10
	tableMinCategoryWidth   = 10

	// Table spacing
	tableColumnSpacing      = 2
	tableColumnGapsExpenses = 3 // Number of gaps between 4 columns
	tableColumnGapsSummary  = 2 // Number of gaps between 3 columns
	tableColumnGapsOverlay  = 2 // Number of gaps between 3 columns

	// Table padding (borders + padding)
	tableBorderPadding = 4

	// Box height calculations
	expensesBoxHeaderRows      = 4 // header + separator + borders
	summaryBoxHeaderRows       = 6 // header + separator + separator + total + borders
	monthlyReportBoxHeaderRows = 4 // header + separator + borders
	overlayHeaderRows          = 6 // header + separator + help + borders

	// Description truncation
	descTruncateSuffix = 3 // "...".length

	// Overlay dimensions (category detail)
	overlayMinWidth          = 60
	overlayMaxWidth          = 100
	overlaySideMargin        = 20 // margin on each side
	overlayMinHeight         = 5
	overlayMaxRows           = 15
	overlayMinHeightFallback = 7

	// Overlay dimensions (confirm delete)
	confirmDeleteOverlayWidth  = 50
	confirmDeleteOverlayHeight = 10

	// Form dimensions
	formWidth          = 30
	promptWidth        = 13
	promptOffsetAmount = 8
	promptOffsetDate   = 6
	promptOffsetType   = 6

	// Row calculation (for update.go)
	titleHeightForRows        = 7
	expensesBoxRowOffset      = 4
	summaryBoxRowOffset       = 6
	monthlyReportBoxRowOffset = 4

	// Box height division
	boxHeightDivisor = 2 // Divide remaining height by 2

	// Border and padding
	borderPadding          = 2
	innerWidthPadding      = 4 // 2 for borders + 2 for padding
	innerHeightPadding     = 2 // top and bottom borders
	borderCornerCharsWidth = 2 // left and right corner characters
)
