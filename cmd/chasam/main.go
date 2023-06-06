package main

/*
 go run main.go di.go --source=/home/martins/Downloads/images/search --target=/home/martins/Downloads/images/benchmark --hash=p-hash --hamming=10
*/

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/progressbar"
)

var (
	cpu      = flag.Int("cpu", runtime.NumCPU(), "--cpu=4")
	source   = flag.String("source", "", "--source=image/source")
	target   = flag.String("target", "", "--target=image/target")
	hashType = flag.String("hash", "d-hash", "--hash=sha1,ed2k,a-hash,m-hash,d-hash,d-hash-v,d-hash-d,p-hash,l-hash")
	hamming  = flag.Int("hamming", 10, "--hamming=10")

	_hashMap      map[hash.Type]bool
	_hashArray    []hash.Type
	provider      = CreateProvider()
	_repository   media.Repository
	countFileCh   = make(chan struct{})
	countMatchCh  = make(chan struct{})
	extractFileCh = make(chan string)

	_extractionFolderPath = "extracted"

	_csv *csv.Writer
)

func main() {
	flag.Parse()

	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		cancelFunc()
	}()

	//ctx, cancelFun := context.WithCancel(context.Background())
	//go func() {
	//	os.Stdin.Read(make([]byte, 1)) // read a single byte
	//	cancelFun()
	//}()

	if *source == "" || *target == "" {
		printHelper()
		os.Exit(0)
	}

	if err := createExtractionFolder(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] Falha ao criar a pasta de extração. Error: %v", err.Error())
		os.Exit(1)
	}

	csvName := fmt.Sprintf(
		"match_%s.csv",
		time.Now().Format("2006-01-02"),
	)
	csvFile, err := os.Create(csvName)
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err.Error())
	}
	_csv = csv.NewWriter(csvFile)
	defer func() {
		_csv.Flush()
		csvFile.Close()
	}()
	_csv.Write([]string{"ORIGEM", "ALVO", "ALVO PATH", "TIPO DO HASH", "HAMMING"})

	printBanner()

	var bar progressbar.Bar
	bar.NewOption("=")

	countFile := 0
	go func() {
		for range countFileCh {
			countFile++
			bar.Play(int64(countFile))
		}

		bar.Finish()
	}()

	countMatch := 0
	go func() {
		for range countMatchCh {
			countMatch++
			bar.Match()
		}
	}()

	go func() {
		for path := range extractFileCh {
			if err := extractFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "[!] Falha ao extrair o arquivo `%s`. Error: %v\n",
					path, err.Error())
			}
		}
	}()

	start := time.Now()

	if err := runMediaSearch(ctx); err != nil {
		fmt.Printf("[!] Error: %v\n", err.Error())
	}

	elapsed := time.Since(start)
	time.Sleep(time.Millisecond * 500)

	color.Printf("\n[>] Pesquisa concluída em: <green>%s</>\n", elapsed)
	color.Printf("[>] Total de arquivos analisados: <green>%d</>\n", countFile)
	color.Printf("[>] Total de match: <green>%d</>\n", countMatch)
	color.Printf("[>] Arquivo de match: <green>%s</>\n", csvFile.Name())

	//panic(fmt.Errorf("%s", "error goroutines"))
}

func createExtractionFolder() error {
	_, err := os.Stat(_extractionFolderPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(_extractionFolderPath, 0o775); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}
	return nil
}

func runMediaSearch(ctx context.Context) error {
	root := *target
	poolSize := *cpu
	_hashArray, _hashMap = makeHashTypes()

	repo, err := provider.MediaRepositoryMem(*source, _hashArray)
	if err != nil {
		return err
	}
	_repository = repo

	s := media.NewSearch(
		ctx,
		root,
		_hashArray,
		onError,
		onSearch,
		onMatch,
		poolSize,
	)
	s.Run()

	close(countFileCh)
	close(extractFileCh)
	close(countMatchCh)
	return nil
}

func makeHashTypes() ([]hash.Type, map[hash.Type]bool) {
	hashMap := make(map[hash.Type]bool)
	var hashArray []hash.Type
	hTypes := strings.Split(*hashType, ",")

	for _, ht := range hTypes {
		switch strings.ToLower(ht) {
		case "sha1":
			hashArray = append(hashArray, hash.SHA1)
			hashMap[hash.SHA1] = true
		case "ed2k":
			hashArray = append(hashArray, hash.ED2K)
			hashMap[hash.ED2K] = true
		case "a-hash":
			hashArray = append(hashArray, hash.AHash)
			hashMap[hash.AHash] = true
		case "m-hash":
			hashArray = append(hashArray, hash.MHash)
			hashMap[hash.MHash] = true
		case "d-hash":
			hashArray = append(hashArray, hash.DHash)
			hashMap[hash.DHash] = true
		case "d-hash-v":
			hashArray = append(hashArray, hash.DHashV)
			hashMap[hash.DHashV] = true
		case "d-hash-d":
			hashArray = append(hashArray, hash.DHashD)
			hashMap[hash.DHashD] = true
		case "p-hash":
			hashArray = append(hashArray, hash.PHash)
			hashMap[hash.PHash] = true
		case "l-hash":
			hashArray = append(hashArray, hash.LHash)
			hashMap[hash.LHash] = true
		case "w-hash":
			hashArray = append(hashArray, hash.WHash)
			hashMap[hash.WHash] = true
		}
	}

	return hashArray, hashMap
}

func onError(_ context.Context, err error) {
	fmt.Fprintf(os.Stderr, "[!] Error: %v\n", err.Error())
}

func onSearch(_ context.Context, m *media.Media) (bool, error) {
	if m.Type() != "image" {
		return false, nil
	}

	countFileCh <- struct{}{}

	if _, ok := _hashMap[hash.SHA1]; ok {
		if src := _repository.FindByHash(hash.SHA1, m.SHA1()); src != "-1" {
			m.AddMatch(src, hash.SHA1.String(), 0)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.ED2K]; ok {
		if src := _repository.FindByHash(hash.ED2K, m.ED2K()); src != "-1" {
			m.AddMatch(src, hash.ED2K.String(), 0)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.AHash]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.AHash, m.AHash(), *hamming); dist != -1 {
			m.AddMatch(src, hash.AHash.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.MHash]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.MHash, m.MHash(), *hamming); dist != -1 {
			m.AddMatch(src, hash.MHash.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.DHash]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.DHash, m.DHash(), *hamming); dist != -1 {
			m.AddMatch(src, hash.DHash.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.DHashV]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.DHashV, m.DHashV(), *hamming); dist != -1 {
			m.AddMatch(src, hash.DHashV.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.DHashD]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.DHashD, m.DHashD(), *hamming); dist != -1 {
			m.AddMatch(src, hash.DHashD.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.PHash]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.PHash, m.PHash(), *hamming); dist != -1 {
			m.AddMatch(src, hash.PHash.String(), dist)
			return true, nil
		}
	}

	if _, ok := _hashMap[hash.LHash]; ok {
		if dist, src := _repository.FindByPerceptualHash(hash.LHash, m.LHash(), *hamming); dist != -1 {
			m.AddMatch(src, hash.LHash.String(), dist)
			return true, nil
		}
	}

	return false, nil
}

func onMatch(_ context.Context, m *media.Media) {
	for _, match := range m.Match() {
		printMatch(match.Name, m.Name(), m.Path(), match.HashType, match.Distance)
	}

	extractFileCh <- m.Path()
	countMatchCh <- struct{}{}
}

func extractFile(path string) error {
	dirname, filename := filepath.Split(path)

	if dirname != "" {
		aux := strings.Split(dirname, ":")
		if len(aux) > 1 {
			dirname = aux[1]
		}
	}

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dstDir := filepath.Join(_extractionFolderPath, dirname)
	if err = os.MkdirAll(dstDir, 0o775); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	dstPath := filepath.Join(dstDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func printMatch(sourceName, targetName, targetPath, hashType string, distHamming int) {
	_csv.Write([]string{
		sourceName,
		targetName,
		targetPath,
		hashType,
		strconv.Itoa(distHamming),
	})
}

func printBanner() {
	fmt.Println("###############################################################################")
	fmt.Printf("#%78s\n", "#")
	color.Printf("%-35s <yellow>%s</> %36s\n", "#", "ChaSAM", "#")
	fmt.Printf("#%78s\n", "#")
	fmt.Println("###############################################################################")
	color.Println("[>] Para abortar pressione as teclas <red>ctrl+c</>")
	color.Println("[>] Iniciando a busca...\n")
}

const templateHelperStr = "\t%-10s \t\t %s\n"

func printHelper() {
	fmt.Println("Uso: chasam --source=images/source --target=images/target --hash=d-hash,d-hash-v --hamming=10")
	fmt.Println("Realiza uma pesquisa de imagens através da comparação de hashs.")

	fmt.Printf("\nArgumentos.\n")
	fmt.Printf(templateHelperStr, "--cpu", "definir o número de núcleos da cpu para o processamento dos hashs")
	fmt.Printf(templateHelperStr, "--hamming", "distância limite entre dois hashs perceptivos")
	fmt.Printf(templateHelperStr, "--hash", "tipo do hash "+
		"(pode ser informado mais de um tipo separados por vírgula)")

	fmt.Printf(templateHelperStr, "\tsha1", "função hash criptográfica de 160 bits")
	fmt.Printf(templateHelperStr, "\ted2k", "hash usado em compartilhamento de arquivos eDonkey")

	fmt.Printf(templateHelperStr, "\ta-hash", "hash médio "+
		"(calculado pela média de todos os valores de cinza da imagem)")

	fmt.Printf(templateHelperStr, "\tm-hash", "hash moda "+
		"(calculado pela moda de todos os valores de cinza da imagem)")

	fmt.Printf(templateHelperStr, "\td-hash", "hash de diferença "+
		"(calcula a diferença entre um pixel e seu vizinho da direita, seguindo o degrade horizontal)")

	fmt.Printf(templateHelperStr, "\td-hash-v", "hash de diferença vertical "+
		"(calcula a diferença entre um pixel e seu vizinho abaixo, seguindo o degrade vertical)")

	fmt.Printf(templateHelperStr, "\td-hash-d", "hash de diferença diagonal "+
		"(calcula a diferença entre um pixel e seu vizinho abaixo e ao lado, seguindo o degrade diagonal)")

	fmt.Printf(templateHelperStr, "\tp-hash", "perceptual hash (calcula aplicando uma transformada discreta de cosseno)")

	fmt.Printf(templateHelperStr, "\tl-hash", "leonard hash (converte a imagem em treshold e calcula aplicando uma transformada discreta de cosseno)")

	fmt.Printf(templateHelperStr, "\tw-hash", "wavelet hash (calcula aplicando uma transformada wavelet bidimensional)")

	fmt.Printf(templateHelperStr, "--source", "diretório de origem com as imagens/vídeos a serem pesquisados")

	fmt.Printf(templateHelperStr, "--target", "diretório alvo onde será realizada a pesquisa por imagens/vídeos")
}
