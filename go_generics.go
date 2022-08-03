# Inhalt und Autor

## Inhaltszusammenfassung

Mit der Version 1.18 hat Go nach jahrelanger Entwicklungszeit Unterstützung von Generics erhalten. In diesem Artikel wird anhand von Beispielen erklärt, wie Generics in Go funktionieren und welche Auswirkungen die neue Sprachfunktion auf Go-Code haben.

## Über den Autor (lang)

Rainer Stropek ist seit über 25 Jahren als Unternehmer in der IT-Industrie tätig. Er gründete und führte in dieser Zeit mehrere IT-Dienstleistungsunternehmen und entwickelt neben seiner Tätigkeit als Trainer und Berater in seiner Firma *software architects* mit seinem Team die preisgekrönte Software *time cockpit* ([https://www.timecockpit.com](https://www.timecockpit.com)).

Rainer hat Abschlüsse an der höheren, technischen Schule für Informatik, Leonding (AT) sowie der University of Derby (UK). Er ist Autor mehrerer Fachbücher und Artikel in Magazinen im Umfeld von Microsoft .NET und C#, Azure, Go und Rust. Seine technischen Schwerpunkte sind Cloud Computing, die Entwicklung verteilter Systeme sowie Datenbanksysteme. Rainer tritt regelmäßig als Speaker und Trainer auf namhaften Konferenzen in Europa und den USA auf. 2010 wurde Rainer von Microsoft zu einem der ersten MVPs für die Azure-Plattform ernannt. Seit 2015 ist Rainer Microsoft Regional Director. 2016 hat Rainer zusätzlich den MVP Award für Visual Studio und Developer Technologies erhalten.

## Über den Autor (kurz)

Rainer Stropek ist IT-Unternehmer, Softwareentwickler, Trainer, Autor und Vortragender im Microsoft-Umfeld. Er ist seit 2010 MVP für Microsoft Azure und entwickelt mit seinem Team die Zeiterfassung für Dienstleistungsprofis *time cockpit* ([https://www.timecockpit.com](https://www.timecockpit.com)).

## Bild des Autors

![Rainer Stropek](https://cddataexchange.blob.core.windows.net/data-exchange/Rainer_Stropek.jpg)
//////////////////////////////////////////////////
// Listing 1: lentil-Struktur als Ausgangsbasis //
//////////////////////////////////////////////////

type lentil struct {
    isGood bool
}

func (l lentil) shouldEat() bool { return !l.isGood }


///////////////////////////////////////////////////////////////
// Listing 2: snail-Struktur mit gleicher Methode wie lentil //
/////////////////////////////////////////////////////////////// 

type snail struct {
    hasHouse bool
}

func (s snail) shouldEat() bool { return !s.hasHouse }


/////////////////////////////////////////////////////////////////////
// Listing 3: Gemeinsames Interface eatOrKeep für lentil und snail //
///////////////////////////////////////////////////////////////////// 

type eatOrKeep interface {
    // Returns true if item should be eaten
    shouldEat() bool
}


///////////////////////////////////////////////////////////////
// Listing 4: Methode zum Verarbeiten von eatOrKeep-Objekten //
///////////////////////////////////////////////////////////////

type bird struct{}

// Removes all items from the slice that should be eaten.
func (p bird) process(items []eatOrKeep) []eatOrKeep {
    result := []eatOrKeep{}
    for _, item := range items {
        if !item.shouldEat() {
            result = append(result, item)
        }
    }

    return result
}

func main() {
    items := []eatOrKeep{
        lentil{isGood: true},
        lentil{isGood: false},
        snail{hasHouse: true},
        snail{hasHouse: false},
    }
    processedItems := bird{}.process(items)
    fmt.Println("Eaten:", len(items)-len(processedItems), "Kept:", len(processedItems))
}


/////////////////////////////////////////////////////////////////////
// Listing 5: Allgemein verwendbare Implementierung mit Reflection //
/////////////////////////////////////////////////////////////////////

// Reflection-based implementation of the filtering.
//
// First parameter must be the slice to be filtered. Second parameter
// must be the filter function returning boolean. Function removes all items
// from slice for which filter function returns false.
func processInterface(itemsSlice interface{}, filter interface{}) interface{} {
    // Get reflection value for slice to process. Note that error handling
    // (e.g. if itemsSlice is not a slice) is not implemented to keep example simple
    // as this article is not a tutorial on Go Runtime Reflection.
    items := reflect.ValueOf(itemsSlice)

    // Create slice to store results
    result := reflect.MakeSlice(items.Type(), 0, 0)

    // Get reflection value for filter function.
    funcValue := reflect.ValueOf(filter)

    // Iterate over all items
    for i := 0; i < items.Len(); i++ {
        // Get item on index i using reflection
        item := items.Index(i)

        // Call filter function using reflection and check result.
        keep := funcValue.Call([]reflect.Value{item})
        if keep[0].Interface().(bool) {
            // Append item to result slice using reflection
            result = reflect.Append(result, item)
        }
    }

    // Return result slice as interface{}
    return result.Interface()
}

func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */
    interfacedItems := processInterface(items, func(item eatOrKeep) bool { return !item.shouldEat() })
    processedItems = interfacedItems.([]eatOrKeep)
    fmt.Println("Eaten:", len(items)-len(processedItems), "Kept:", len(processedItems))
}


///////////////////////////////////
// Listing 6: Generische Methode //
///////////////////////////////////

func process[I any](items []I, filter func(i I) bool) []I {
    result := []I{}
    for _, item := range items {
        if filter(item) {
            result = append(result, item)
        }
    }

    return result
}
func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */
    processedItems = process(items, func(item eatOrKeep) bool { return !item.shouldEat() })
    // You could manually specify type parameter (see next line), not necessary, will lead to warning.
    //processedItems = process[eatOrKeep](items, func(item eatOrKeep) bool { return !item.shouldEat() })
    fmt.Println("Eaten:", len(items)-len(processedItems), "Kept:", len(processedItems))
}


//////////////////////////////////////////////////////////
// Listing 7: Verwaltung einer Collection ohne Generics //
//////////////////////////////////////////////////////////

type itemsGroup struct {
    eatOrKeep
    count int
}

type itemsBag struct {
    bag []itemsGroup
}

func newItemsBag() *itemsBag {
    return &itemsBag{
        bag: make([]itemsGroup, 0),
    }
}

func (b *itemsBag) append(item eatOrKeep) {
    // Check if the new item is identical to the one previously insert.
    if len(b.bag) == 0 || item.shouldEat() != b.bag[len(b.bag)-1].shouldEat() {
        // First or different item, add a new group
        b.bag = append(b.bag, itemsGroup{eatOrKeep: item, count: 1})
    } else {
        // Identical item, increment count
        b.bag[len(b.bag)-1].count++
    }
}

// Recreates the slice of items
func (b itemsBag) getItems() []eatOrKeep {
    result := make([]eatOrKeep, 0)

    // Iterate over all groups
    for _, group := range b.bag {
        // Recreate items based on count
        for i := 0; i < group.count; i++ {
            result = append(result, group.eatOrKeep)
        }
    }

    return result
}

func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */

    bag := newItemsBag()
    bag.append(lentil{isGood: true})
    bag.append(lentil{isGood: true})
    bag.append(lentil{isGood: false})
    bag.append(lentil{isGood: false})
    processedItems = process(bag.getItems(), func(item eatOrKeep) bool { return !item.shouldEat() })
    fmt.Println("Eaten:", len(items)-len(processedItems), "Kept:", len(processedItems))
}


//////////////////////////////////////////////////////
// Listing 8: Umsetzung der Collection mit Generics //
//////////////////////////////////////////////////////

// Stores an element plus its count. Note generic type for `item`
type genericItemsGroup[T any] struct {
    item  T
    count int
}

// Stores a collection of item groups and a generic function used to
// compare two items.
type genericItemsBag[T any] struct {
    bag              []genericItemsGroup[T]
    equalityComparer func(T, T) bool
}

func newGenericItemsBag[T any](comparer func(T, T) bool) *genericItemsBag[T] {
    return &genericItemsBag[T]{
        bag:              make([]genericItemsGroup[T], 0),
        equalityComparer: comparer,
    }
}

func (b *genericItemsBag[T]) append(item T) {
    if len(b.bag) == 0 || !b.equalityComparer(item, b.bag[len(b.bag)-1].item) {
        b.bag = append(b.bag, genericItemsGroup[T]{item: item, count: 1})
    } else {
        b.bag[len(b.bag)-1].count++
    }
}

func (b genericItemsBag[T]) getItems() []T {
    result := make([]T, 0)
    for _, group := range b.bag {
        for i := 0; i < group.count; i++ {
            result = append(result, group.item)
        }
    }

    return result
}

func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */

    // Create generic items bag. Note that we do not need to specify any type parameter.
    // Go can figure them out using type inference.
    genericBag := newGenericItemsBag(func(lhs eatOrKeep, rhs eatOrKeep) bool { return lhs.shouldEat() == rhs.shouldEat() })
    genericBag.append(lentil{isGood: true})
    genericBag.append(lentil{isGood: true})
    genericBag.append(lentil{isGood: false})
    genericBag.append(lentil{isGood: false})
    processedItems = process(genericBag.getItems(), func(item eatOrKeep) bool { return !item.shouldEat() })
    fmt.Println("Eaten:", len(items)-len(processedItems), "Kept:", len(processedItems))
}


////////////////////////////////////
// Listing 9: Generische Channels //
////////////////////////////////////

// Filter channel based on a given, generic filter function
func processChannel[I any](items <-chan I, filter func(i I) bool) <-chan I {
    out := make(chan I)
    go func() {
        defer close(out)
        for item := range items {
            if filter(item) {
                out <- item
            }
        }
    }()
    return out
}

func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */

    // Fill buffered channel with lentils
    in := make(chan eatOrKeep, 4)
    in <- lentil{isGood: true}
    in <- lentil{isGood: true}
    in <- lentil{isGood: false}
    in <- lentil{isGood: false}
    close(in)

    // Use generic channel processing
    total := len(in)
    remaining := 0
    for range processChannel(in, func(item eatOrKeep) bool { return !item.shouldEat() }) {
        remaining++
    }
    fmt.Println("Eaten:", total-remaining, "Kept:", remaining)
}


/////////////////////////////////////////////////
// Listing 10: Vordefiniertes Type Set Ordered //
/////////////////////////////////////////////////

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type Ordered interface {
    Integer | Float | ~string
}


/////////////////////////////////////////////////////
// Listing 11: Erweiterung des eatOrKeep Interface //
/////////////////////////////////////////////////////

const (
    SMALL  = 1
    MEDIUM = 2
    LARGE  = 3
)

// Lentil with size information
type sizedLentil struct {
    lentil
    lentilSize int
}

func (l sizedLentil) size() int { return l.lentilSize }

type sized interface {
    size() int
}

// Define an interface that will be used as a type constraint
type sizedEatOrKeep interface {
    sized
    eatOrKeep
}


////////////////////////////////////////////////
// Listing 12: Anwendung von Type Constraints //
////////////////////////////////////////////////

func processAndSort[I sizedEatOrKeep](items []I, filter func(i I) bool) []I {
    // Filter exactly as before, code omitted to focus on type constraints
    /* ... */

    bubblesort(result, func(item I) int { return item.size() })
    return result
}

func bubblesort[I any, O constraints.Ordered](items []I, toOrdered func(item I) O) {
    for itemCount := len(items) - 1; ; itemCount-- {
        hasChanged := false
        for index := 0; index < itemCount; index++ {
            // We use the provided function to turn each item into a type compatible
            // with Ordered. With that, we can use comparison operators.
            if toOrdered(items[index]) > toOrdered(items[index+1]) {
                items[index], items[index+1] = items[index+1], items[index]
                hasChanged = true
            }
        }
        if !hasChanged {
            break
        }
    }
}

func main() {
    items := []eatOrKeep{ /*...*/ }
    /* ... */

    sizedItems := []sizedEatOrKeep{
        sizedLentil{lentilSize: LARGE, lentil: lentil{isGood: true}},
        sizedLentil{lentilSize: MEDIUM, lentil: lentil{isGood: false}},
        sizedLentil{lentilSize: SMALL, lentil: lentil{isGood: true}},
    }
    processedOrderd := processAndSort(sizedItems, func(item sizedEatOrKeep) bool { return !item.shouldEat() })
    fmt.Println("Eaten:", len(sizedItems)-len(processedOrderd), "Kept:", len(processedOrderd))
    for _, sortedItem := range processedOrderd {
        fmt.Println("Size:", sortedItem.size())
    }
}