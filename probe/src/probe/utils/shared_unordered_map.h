#ifndef SHAREDUNORDERMAP_H
#define SHAREDUNORDERMAP_H

#include <condition_variable>
#include <bits/hashtable.h>
#include <unordered_map>

template<typename T, typename U>
class shared_unordered_map {
public:
	typedef typename std::unordered_map<T, U>::iterator iterator_;

	shared_unordered_map();

	~shared_unordered_map();

	int size();

	bool empty();

	void erase(iterator_);

	void erase(T key);

	iterator_ begin();

	iterator_ end();

	iterator_ find(T key);

	void insert(T key, U value);

	typename std::unordered_map<T, U> *rangeStart();

	void rangeEnd();

private:
	std::condition_variable cond_;
	std::mutex mutex_;
	std::unordered_map<T, U> *unordered_map1;
	std::unordered_map<T, U> *unordered_map2;

};

template<typename T, typename U>
shared_unordered_map<T, U>::shared_unordered_map() {
	unordered_map1 = new std::unordered_map<T, U>();
}

template<typename T, typename U>
shared_unordered_map<T, U>::~shared_unordered_map() = default;

template<typename T, typename U>
int shared_unordered_map<T, U>::size() {
	std::unique_lock<std::mutex> mlock(mutex_);
	int size = unordered_map1->size();
	mlock.unlock();
	return size;
}

template<typename T, typename U>
bool shared_unordered_map<T, U>::empty() {
	return this->size() == 0;
}

template<typename T, typename U>
void shared_unordered_map<T, U>::erase(iterator_ index) {
	std::unique_lock<std::mutex> mlock(mutex_);
	while (unordered_map1->empty()) {
		cond_.wait(mlock);
	}
	unordered_map1->erase(index);
}

template<typename T, typename U>
typename std::unordered_map<T, U>::iterator shared_unordered_map<T, U>::begin() {
	std::unique_lock<std::mutex> mlock(mutex_);
	iterator_ index = unordered_map1->begin();
	mlock.unlock();
	return index;
}

template<typename T, typename U>
typename std::unordered_map<T, U>::iterator shared_unordered_map<T, U>::end() {
	std::unique_lock<std::mutex> mlock(mutex_);
	iterator_ index = unordered_map1->end();
	mlock.unlock();
	return index;
}

template<typename T, typename U>
void shared_unordered_map<T, U>::insert(T key, U value) {
	std::unique_lock<std::mutex> mlock(mutex_);
	unordered_map1->insert({key, value});
	mlock.unlock();
	cond_.notify_one();
}

template<typename T, typename U>
typename std::unordered_map<T, U>::iterator shared_unordered_map<T, U>::find(T key) {
	std::unique_lock<std::mutex> mlock(mutex_);
	typename std::unordered_map<T, U>::iterator index = unordered_map1->find(key);
	mlock.unlock();
	return index;
}

template<typename T, typename U>
void shared_unordered_map<T, U>::rangeEnd() {
	delete unordered_map2;
}

template<typename T, typename U>
typename std::unordered_map<T, U> *shared_unordered_map<T, U>::rangeStart() {
	unordered_map2 = new std::unordered_map<T, U>();
	std::unique_lock<std::mutex> mlock(mutex_);
	iterator_ index = unordered_map1->begin();
	while (index != unordered_map1->end()) {
		unordered_map2->insert({index->first, index->second});
		index++;
	}
	mlock.unlock();
	return unordered_map2;
}

template<typename T, typename U>
void shared_unordered_map<T, U>::erase(T key) {
	std::unique_lock<std::mutex> mlock(mutex_);
	while (unordered_map1->empty()) {
		cond_.wait(mlock);
	}
	unordered_map1->erase(key);
}


#endif //SHAREDUNORDERMAP_H
