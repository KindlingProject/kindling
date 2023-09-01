use crate::cpu_analyzer::model::Segment;

#[derive(Debug)]
pub struct CircleQueue {
    length: usize,
    data: Vec<Segment>,
}

impl CircleQueue {
    pub fn new(length: usize) -> Self {
        CircleQueue {
            length,
            data: vec![Default::default(); length],
        }
    }

    pub fn get_by_index(&self, index: usize) -> Option<&Segment> {
        self.data.get(index)
    }

    pub fn get_by_index_mut(&mut self, index: usize) -> Option<&mut Segment> {
        self.data.get_mut(index)
    }

    pub fn update_by_index(&mut self, index: usize, val: Segment) {
        if index < self.length {
            self.data[index] = val;
        }
    }

    pub fn update_by_moved_index(&mut self, i: usize, moved_index: usize) {
        self.data.swap(i, moved_index);
    } 

    pub fn clear(&mut self) {
        self.data = vec![Default::default(); self.length];
    }
}
