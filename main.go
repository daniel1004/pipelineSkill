package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var bufferSize int = 3

const bufferDrainInterval time.Duration = 10 * time.Second

// Функция логирования для вывода в консоль
func log(message string) {
	fmt.Println(message)
}

func FilterPositive(done <-chan struct{}, inputData <-chan int) <-chan int {
	onlyPositiveData := make(chan int)
	go func() {
		defer close(onlyPositiveData)
		for {
			select {
			case <-done:
				return
			case value, ok := <-inputData:
				if !ok {
					return
				}
				if value > 0 {
					log(fmt.Sprintf("Фильтрация положительных чисел: %d", value))
					select {
					case onlyPositiveData <- value:
						log(fmt.Sprintf("Передано положительное число: %d", value))
					case <-done:
						return
					}
				}
			}
		}
	}()
	return onlyPositiveData
}

func FilterThree(done <-chan struct{}, inputData <-chan int) <-chan int {
	onlyThreeData := make(chan int)
	go func() {
		defer close(onlyThreeData)
		for {
			select {
			case <-done:
				return
			case value, ok := <-inputData:
				if !ok {
					return
				}
				if value%3 == 0 {
					log(fmt.Sprintf("Фильтрация чисел кратных 3: %d", value))
					select {
					case onlyThreeData <- value:
						log(fmt.Sprintf("Переданно число кратное 3: %d", value))
					case <-done:
					}
				}

			}
		}
	}()
	return onlyThreeData
}

func main() {
	inputData := func() (<-chan struct{}, <-chan int) {
		output := make(chan int)
		done := make(chan struct{})
		go func() {
			defer close(done)
			scanner := bufio.NewScanner(os.Stdin)
			var str string
			fmt.Println("Введите число для продолжения")
			log("Введите число для продолжния или stop для завершения")
			for {
				scanner.Scan()
				str = scanner.Text()
				if strings.EqualFold(str, "stop") {
					fmt.Println("Программа завершила работу")
					log("Программа завершила работу")
					close(output)
					return
				}
				val, err := strconv.Atoi(scanner.Text())
				if err != nil {
					fmt.Println("Только int!")
					log("Ошибка: только int!")
					continue
				}
				output <- val
			}
		}()
		return done, output
	}

	buferisation := func(done <-chan struct{}, input <-chan int) <-chan int {
		r := ring.New(bufferSize)
		preR := r
		mu := sync.Mutex{}
		bufferedIntChan := make(chan int)

		go func() {
			defer close(bufferedIntChan)
			for {
				select {
				case <-done:
					return
				case value, ok := <-input:
					if !ok {
						return
					}
					mu.Lock()
					r.Value = value
					log(fmt.Sprintf("Добавлено в буфер: %d", value))
					r = r.Next()
					mu.Unlock()
				}
			}
		}()

		go func() {
			for {
				select {
				case <-done:
					return
				case <-time.After(bufferDrainInterval):
					mu.Lock()
					preR.Do(func(p interface{}) {
						if p != nil {
							select {
							case bufferedIntChan <- p.(int):
								log(fmt.Sprintf("Извлечено из буфера: %d", p.(int)))
							case <-done:
								return
							}
							preR.Value = nil
							preR = preR.Next()

						}
					})
					mu.Unlock()
				}
			}
		}()
		return bufferedIntChan
	}
	potrebitel := func(done <-chan struct{}, input <-chan int) {
		for {
			select {
			case <-done:
				return
			case val := <-input:
				fmt.Printf("Обраюотаны данные: %d\n", val)
				log(fmt.Sprintf("Обработаны данные: %d", val))
			}
		}
	}
	done, output := inputData()
	potrebitel(done, buferisation(done, FilterThree(done, FilterPositive(done, output))))

}
