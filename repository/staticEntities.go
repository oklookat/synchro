package repository

var (
	EntityAlbum  = newEntityRepository("album", "linked_album")
	EntityArtist = newEntityRepository("artist", "linked_artist")
	EntityTrack  = newEntityRepository("track", "linked_track")
)

func NewLinkableEntityAlbum(rem *Service) *LinkableEntity {
	return newLinkableEntity("linked_album", rem)
}

func NewLinkableEntityArtist(rem *Service) *LinkableEntity {
	return newLinkableEntity("linked_artist", rem)
}

func NewLinkableEntityTrack(rem *Service) *LinkableEntity {
	return newLinkableEntity("linked_track", rem)
}
