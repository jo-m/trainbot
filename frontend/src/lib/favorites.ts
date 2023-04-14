import { defineStore, type StateTree } from 'pinia'

interface State {
  favorites: Set<number>
}

const useFavoritesStore = defineStore('favorites', {
  state: (): State => {
    return {
      favorites: new Set()
    }
  },
  getters: {
    isFav: (state: StateTree) => {
      const s = state as State
      return (id: number) => s.favorites.has(id)
    }
  },
  actions: {
    toggle(id: number) {
      if (this.favorites.has(id)) {
        this.favorites.delete(id)
      } else {
        this.favorites.add(id)
      }
    },
    clear() {
      this.favorites.clear()
    }
  },
  persist: {
    storage: localStorage,
    paths: ['favorites'],
    serializer: {
      serialize: (state: StateTree): string => {
        // Convert set to Array for storage.
        const s = state as State
        return JSON.stringify(Array.from(s.favorites))
      },
      deserialize: (state: string): StateTree => {
        // Array to Set.
        try {
          return { favorites: new Set(JSON.parse(state)) }
        } catch {
          return { favorites: new Set() }
        }
      }
    }
  }
})

export default useFavoritesStore
