package scenes

type LaneId uint

const (
	BassLaneId LaneId = iota
	GuitarLaneId
	DrumsLaneId
)

type Note struct {
	Lane      LaneId  // Lajur tempat not ini berada (0 hingga laneCount-1).
	Tick      float64 // Waktu (dalam "tick") kapan not ini harusnya ditekan.
	IsActive  bool    // Status apakah not ini masih dalam permainan (belum ditekan atau terlewat).
	YPosition float64 // Posisi Y not di layar saat ini.
}

func NewNoteChart() []Note {
	return []Note{
		{
			Lane:     BassLaneId,
			Tick:     0,
			IsActive: true,
		},
	}
}
