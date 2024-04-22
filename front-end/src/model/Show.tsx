export interface Show {
	title: string
	year: number
	id: string
	rating: number
	isSelected: boolean
	onSelect: () => void;
}