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
	cachedList                  []model.ContentCommand
	cachedMapByAlias            map[string]model.ContentCommand
	cachedMapByID               map[int]model.ContentCommand
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

	var csvContentCommands []model.CsvContentCommand
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

	err = csvutil.Unmarshal(csvFile, &csvContentCommands)
	if err != nil {
		logging.Log.Error(
			logPackage,
			"CsvObjectStorageCachedContentCommandRepository.refreshCache",
			err,
			"csv file parsing error",
		)
		return err
	}

	var list []model.ContentCommand
	for _, csv := range csvContentCommands {
		command := convertCsvContentCommand(&csv)
		for _, alias := range command.GetAliases() {
			repo.maxCommandAliasStringLength = utils.MaxInt(repo.maxCommandAliasStringLength, len(alias))
		}
		list = append(list, command)
	}

	var mapByID = make(map[int]model.ContentCommand, len(list))
	for _, command := range list {
		mapByID[command.GetID()] = command
	}

	var mapByAlias = make(map[string]model.ContentCommand, len(list))
	for _, command := range list {
		for _, alias := range command.GetAliases() {
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

func (repo *CsvObjectStorageCachedContentCommandRepository) FindAll() []model.ContentCommand {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return []model.ContentCommand{}
		}
	}

	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedList)))
	if ptr != nil {
		commands := *(*[]model.ContentCommand)(ptr)
		if len(commands) != 0 {
			return commands
		}
	}

	return []model.ContentCommand{}
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindByCommandAlias(alias string) *model.ContentCommand {
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
		commands := *(*map[string]model.ContentCommand)(ptr)
		if commandByAlias, exists := commands[alias]; exists {
			return &commandByAlias
		}
	}
	return nil
}

func (repo *CsvObjectStorageCachedContentCommandRepository) FindById(ID int) *model.ContentCommand {
	if repo.isNeededInvalidateCache() {
		err := repo.refreshCache()
		if err != nil {
			return nil
		}
	}

	contentSourcesPtr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&repo.cachedMapByID)))
	if contentSourcesPtr != nil {
		commands := *(*map[int]model.ContentCommand)(contentSourcesPtr)
		if commandById, exists := commands[ID]; exists {
			return &commandById
		}
	}
	return nil
}

func convertCsvContentCommand(csv *model.CsvContentCommand) model.ContentCommand {
	var types []model.MediaContentType
	for _, rawType := range strings.Split(csv.MediaContentTypes, ",") {
		types = append(types, model.MediaContentType(rawType))
	}

	return model.NewContentCommand(
		csv.ID,
		strings.Split(csv.Commands, ","),
		types,
		strings.Split(csv.CommunityIDs, ","),
	)
}
