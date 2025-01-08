export interface Show {
  movieId: number
	title: string
	year: number
	jellyfinId: string
	rating: number
	isSelected: boolean
  imageData: Uint8Array
	onSelect: () => void;
}
