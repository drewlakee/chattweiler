package storage

import (
	"chattweiler/internal/logging"
	"chattweiler/internal/repository/model"
	"chattweiler/internal/utils"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/csvutil"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type CsvObjectStorageCachedContentCommandRepository struct {
	client               *s3.Client
	bucket               string
	key                  string
	cacheRefreshInterval time.Duration
	lastCacheRefresh     time.Time
	refreshMutex         sync.Mutex

	maxCommandAliasStringLength int
	cachedList                  []model.Command
	cachedMapByAlias            map[string]model.Command
	cachedMapByID               map[int]model.Command
}

func NewCsvObjectStorageCachedContentSourceRepository(client *s3.Client, bucket, key string, cacheRefreshInterval time.Duration) *CsvObjectStorageCachedContentCommandRepository {
	repository := CsvObjectStorageCachedContentCommandRepository{
		client:               client,
		bucket:               bucket,
		key:                  key,
		cacheRefreshInterval: cacheRefreshInterval,
		lastCacheRefresh:     time.Now(),
		cachedList:           nil,
		cachedMapByAlias:     nil,
	}
	err := repository.refreshCache()
	if err != nil {
		panic(err)
	}
	return &repository
}

func (repo *CsvObjectStorageCachedContentCommandRepository) isNeededInvalidateCache() bool {
	return time.Now().After(repo.lastCacheRefresh.Add(repo.cacheRefreshInterval))
}

func (repo *CsvObjectStorageCachedContentCommandRepository) refreshCache() error {
	startTime := time.Now().UnixMilli()

	// cache refresh lock
	repo.refreshMutex.Lock()
	defer repo.refreshMutex.Unlock()

	object, err := repo.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &repo.bucket,
		Key:    &repo.key,
	})
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageCachedContentCommandRepository.refreshCache",
			err,
			"s3 client error: bucket - %s, key - %s", repo.bucket, repo.key,
		)
		return err
	}

	var csvCommands []model.CsvCommand
	csvFile, err := io.ReadAll(object.Body)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageCachedContentCommandRepository.refreshCache",
			err,
			"csv file reading error",
		)
		return err
	}

	err = csvutil.Unmarshal(csvFile, &csvCommands)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageCachedContentCommandRepository.refreshCache",
			err,
			"csv file parsing error",
		)
		return err
	}

	var list []model.Command
	for _, csv := range csvCommands {
		command := convertCsvContentCommand(&csv)
		for _, alias := range command.Aliases {
			repo.maxCommandAliasStringLength = utils.MaxInt(repo.maxCommandAliasStringLength, len(alias))
		}
		list = append(list, command)
	}

	var mapByID = make(map[int]model.Command, len(list))
	for _, command := range list {
		mapByID[command.ID] = command
	}

	var mapByAlias = make(map[string]model.Command, len(list))
	for _, command := range list {
		for _, alias := range command.Aliases {
			mapByAlias[strings.ToLower(alias)] = command
		}
	}

	listPtr := unsafe.Pointer(&list)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedList)), listPtr)

	mapByAliasPtr := unsafe.Pointer(&mapByAlias)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedMapByAlias)), mapByAliasPtr)

	mapByIdPtr := unsafe.Pointer(&mapByID)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedMapByID)), mapByIdPtr)

	repo.lastCacheRefresh = time.Now()
	logging.Log.Info(logPackage, "CsvObjectStorageCachedContentCommandRepository.refreshCache", "Cache successfully updated for %d ms", time.Now().UnixMilli()-startTime)
	return nil
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindAll() []model.Command {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return []model.Command{}
		}
	}

	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedList)))
	if ptr != nil {
		commands := *(*[]model.Command)(ptr)
		if len(commands) != 0 {
			return commands
		}
	}

	return []model.Command{}
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindByCommandAlias(alias string) *model.Command {
	if len(alias) > repo.maxCommandAliasStringLength {
		return nil
	}

	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return nil
		}
	}

	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedMapByAlias)))
	if ptr != nil {
		commands := *(*map[string]model.Command)(ptr)
		if commandByAlias, exists := commands[strings.ToLower(alias)]; exists {
			return &commandByAlias
		}
	}
	return nil
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindById(ID int) *model.Command {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return nil
		}
	}

	contentSourcesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedMapByID)))
	if contentSourcesPtr != nil {
		commands := *(*map[int]model.Command)(contentSourcesPtr)
		if commandById, exists := commands[ID]; exists {
			return &commandById
		}
	}
	return nil
}

func convertCsvContentCommand(csv *model.CsvCommand) model.Command {
	var types []model.MediaContentType
	for _, rawType := range strings.Split(csv.MediaContentTypes, ",") {
		types = append(types, model.MediaContentType(rawType))
	}

	return model.NewCommand(
		csv.ID,
		csv.Type,
		strings.Split(csv.Commands, ","),
		types,
		strings.Split(csv.CommunityIDs, ","),
	)
}
